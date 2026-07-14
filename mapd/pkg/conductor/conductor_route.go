// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conductor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/avi"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	Registry.Register("route", DispatcherFactory(
		mapd.EventEmitterRoute,
		func() error {
			var routes []model.Route
			if err := mapd.Database().Select("id").Where("`provision_phase` < ? OR `provisioned` != ?", model.RouteProvisionNone, model.RouteProvisionFinish).Find(&routes).Error; err != nil {
				return fmt.Errorf("error obtaining existing routes: %v", err)
			}
			go func() {
				for _, each := range routes {
					mapd.EventEmitterRoute.Next(&mapd.Event{
						Type: mapd.EventTypeAdd,
						OID:  each.ID,
					})
				}
			}()
			return nil
		},
		syncRoute,
	))
}

// syncRoute handle Route transactions.
func syncRoute(id uuid.UUID, log *zap.SugaredLogger) time.Duration {
	defer log.Sync()
	next := time.Duration(-1)

	route := &model.Route{ID: id}
	finished := false
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(route).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				finished = true
				return nil
			}
			return err
		}

		if route.ProvisionPhase > model.RouteProvisionNone && route.Provisioned == model.RouteProvisionFinish {
			finished = true
			return nil
		}

		lb := &model.LoadBalancer{ID: route.LoadBalancerID.UUID}
		if err := tx.Take(lb).Error; err != nil {
			return err
		}

		block, err := cryptoutil.NewAESGCMCipher(lb.ID[:]) // AES-128
		if err != nil {
			panic(err)
		}

		ac, err := avi.NewAVI(lb.Endpoint, lb.AuthUser, block.Decrypt(lb.AuthPassword), lb.InsecureTLS, lb.CACert)
		if err != nil {
			panic(err)
		}

		switch lb.Type {
		case model.LoadBalancerTypeAVI:
			spec := &model.RouteAVISpec{}
			if err := json.Unmarshal(route.Spec, spec); err != nil {
				panic(err)
			}

			if route.ProvisionPhase < model.RouteProvisionNone {
				log.Infof("route %s is about to be deleted.", route.Name)
				if spec.VSID != "" {
					if err := ac.DeleteVirtualService(spec.VSID); err != nil && !avi.IsErrObjectNotExists(err) {
						log.Errorf("failed to delete virtual service of route %s due to: %v.", route.Name, err)
						return nil
					}
				}
				if spec.VSVIPID != "" && !spec.PersistentVSVIP {
					if err := ac.DeleteVSVIP(spec.VSVIPID); err != nil && !avi.IsErrObjectNotExists(err) {
						log.Errorf("failed to delete vsvip of route %s due to: %v.", route.Name, err)
						return nil
					}
				}
				if spec.PoolID != "" {
					if err := ac.DeletePool(spec.PoolID); err != nil && !avi.IsErrObjectNotExists(err) {
						log.Errorf("failed to delete pool of route %s due to: %v.", route.Name, err)
						return nil
					}
				}
				if err := tx.Where("`route_id` = ?", route.ID).Delete(&model.RouteBackend{}).Error; err != nil {
					return err
				}
				if err := tx.Delete(route).Error; err != nil {
					return err
				}
				log.Infof("route %s has been successfully deleted.", route.Name)
				finished = true
				return nil
			}

			if route.ProvisionPhase < model.RouteProvisionFinish {
				log = log.Named("phase").Named(strconv.FormatInt(int64(route.ProvisionPhase), 10))
			}

			switch route.ProvisionPhase {
			case model.RouteProvisionPrepare:
				log.Infof("provisioning pools for route %s...", route.Name)
				var backends []model.RouteBackend
				if err := tx.Where("`route_id` = ?", route.ID).Find(&backends).Error; err != nil {
					return err
				}

				poolName := "map-pool-" + standardizeString(route.Name) + "-" + route.OrganizationID.String()
				pool := &avi.Pool{
					Name:                            poolName,
					LBAlgorithm:                     spec.LBAlgorithm,
					LBAlgorithmHash:                 spec.LBAlgorithmHash,
					LBAlgorithmConsistentHashHeader: spec.LBAlgorithmConsistentHashHeader,
					LBAlgorithmCoreNonaffinity:      spec.LBAlgorithmCoreNonaffinity,
					ConnPoolProperties: &avi.PoolConnPoolProperties{
						UpstreamConnpoolConnIdleTmo:    60000,
						UpstreamConnpoolConnLifeTmo:    600000,
						UpstreamConnpoolConnMaxReuse:   0,
						UpstreamConnpoolServerMaxCache: 0,
					},
				}

				for _, each := range backends {
					ip := net.ParseIP(each.IP)
					var typ string
					switch len(ip) {
					case net.IPv4len:
						typ = "V4"
					case net.IPv6len:
						typ = "V6"
					}
					pool.Servers = append(pool.Servers, &avi.Server{
						Enabled: true,
						IP: &avi.IPAddr{
							Addr: each.IP,
							Type: typ,
						},
						Port:  each.Port,
						Ratio: each.Ratio,
					})
				}
				if len(pool.Servers) > 0 {
					pool.DefaultServerPort = pool.Servers[0].Port
					pool.MinHealthMonitorsUp = 1
					pool.MinServersUp = 1
				} else {
					pool.DefaultServerPort = 8080
				}

				if spec.PoolID != "" {
					pool.UUID = spec.PoolID
					log.Infof("updating pool %s [%s]...", route.Name, poolName)
					if _, err = ac.UpdatePool(pool); err != nil {
						log.Errorf("failed to update pool of route %s due to: %v.", route.Name, err)
						return nil
					}
				} else {
					log.Infof("creating pool %s [%s]...", route.Name, poolName)
					if pool, err = ac.CreatePool(pool); err != nil {
						log.Errorf("failed to create pool of route %s due to: %v.", route.Name, err)
						return nil
					}
					spec.PoolID = pool.UUID
				}

				updates := make(map[string]interface{})
				updates["spec"], err = json.Marshal(spec)
				if err != nil {
					panic(err)
				}
				updates["provisioned"] = model.RouteProvisionPrepare
				updates["provision_phase"] = model.RouteProvisionIP
				if err := tx.Model(route).Updates(updates).Error; err != nil {
					return err
				}
				log.Infof("provisioned pools of route %s. Pending next phase...", route.Name)
				next = time.Second

			case model.RouteProvisionIP:
				log.Infof("provisioning IP address for route %s...", route.Name)
				var vsvip *avi.VSVIP
				if spec.VSVIPID != "" {
					vsvip, err = ac.GetVSVIP(spec.VSVIPID)
					if err != nil {
						log.Errorf("failed to obtain VsVip of route %s due to: %v.", route.Name, err)
						return nil
					}
				} else {
					lbSpec := &model.LoadBalancerAVISpec{}
					if err := json.Unmarshal(lb.Spec, lbSpec); err != nil {
						panic(err)
					}

					vip := &avi.VIP{
						Enabled:                true,
						AutoAllocateIP:         true,
						AutoAllocateFloatingIP: false,
						AVIAllocatedVIP:        false,
						AVIAllocatedFIP:        false,
						AutoAllocateIPType:     "V4_ONLY",
						PrefixLength:           32,
					}
					if lbSpec.PreferredNetwork != "" {
						log.Infof("using network %s as a dedicated network on load balancer %s. Attempt to allocate VsVip on network %s.", lbSpec.PreferredNetwork, strconv.Quote(lb.Name), lbSpec.PreferredNetwork)
						network, err := findNetwork(ac, lbSpec.PreferredNetwork)
						if err != nil {
							log.Errorf("failed to obtain network from load balancer %s due to: %v.", strconv.Quote(lb.Name), err)
							return nil
						}
						if network != nil {
							log.Infof("found dedicated network %s: %s.", network.Name, network.UUID)
							vip.NetworkRef = network.URL

							log.Infof("picking up available IP range on network %s from load balancer %s...", lbSpec.PreferredNetwork, strconv.Quote(lb.Name))
							networkRuntime, err := ac.GetNetworkRuntime(network.UUID)
							if err != nil {
								log.Errorf("failed to obtain network runtime from load balancer %s due to: %v.", strconv.Quote(lb.Name), err)
								return nil
							}
						LOOP:
							for _, subnetRuntime := range networkRuntime.SubnetRuntime {
								for _, ipRangeRuntime := range subnetRuntime.IPRangeRuntimes {
									if ipRangeRuntime.FreeIPCount > 0 && ipRangeRuntime.Type != "STATIC_IPS_FOR_SE" {
										switch subnetRuntime.Prefix.IPAddr.Type {
										case "V4":
											vip.Subnet = &subnetRuntime.Prefix
											break LOOP
										default:
											// we currently do not support IPv6
										}
									}
								}
							}
							if vip.Subnet == nil && vip.Subnet6 == nil {
								log.Errorf("no available subnet for creating VsVip on network %s from load balancer %s.", lbSpec.PreferredNetwork, strconv.Quote(lb.Name))
								return nil
							}
							vip.IPAMNetworkSubnet = &avi.IPNetworkSubnet{
								NetworkRef: vip.NetworkRef,
								Subnet:     vip.Subnet,
								Subnet6:    vip.Subnet6,
							}
						} else {
							log.Errorf("unable to find preferred network %s. Attempt to proceed the provision later.", lbSpec.PreferredNetwork)
							return nil
						}
					} else {
						log.Infof("no preferred network was specified on load balancer %s. Attempt to allocate VsVip from global network.", strconv.Quote(lb.Name))
					}
					vsVip := &avi.VSVIP{
						Name: "map-vsvip-" + standardizeString(route.Name) + "-" + route.OrganizationID.String(),
						VIP:  []*avi.VIP{vip},
					}

					vsvip, err = ac.CreateVSVIP(vsVip)
					if err != nil {
						if !avi.IsErrObjectAlreadyExists(err) {
							log.Errorf("failed to create VsVip of route %s due to: %v.", route.Name, err)
							return nil
						}
						v, e := findVSVIP(ac, vsVip.Name)
						if e != nil {
							log.Errorf("failed to create VsVip of route %s due to: %v.", route.Name, err)
							return nil
						}
						vsvip = v
					}
				}
				log.Infof("allocated VsVip %s to route %s.", vsvip.Name, route.Name)
				updates := make(map[string]interface{})
				spec.VSVIPID = vsvip.UUID
				if len(vsvip.VIP) < 1 || vsvip.VIP[0].IPAddress == nil {
					log.Errorf("no available IP address was found on VsVip %s.", vsvip.Name)
					updates := make(map[string]interface{})
					updates["spec"], err = json.Marshal(spec)
					if err != nil {
						panic(err)
					}
					if err := tx.Model(route).Updates(updates).Error; err != nil {
						return err
					}
					return nil
				}

				updates["spec"], err = json.Marshal(spec)
				if err != nil {
					panic(err)
				}
				updates["ip"] = vsvip.VIP[0].IPAddress.Addr
				updates["provisioned"] = model.RouteProvisionIP
				updates["provision_phase"] = model.RouteProvisionRoute
				if err := tx.Model(route).Updates(updates).Error; err != nil {
					return err
				}
				log.Infof("provisioned VsVip of route %s. Pending next phase...", route.Name)
				next = time.Second

			case model.RouteProvisionRoute:
				log.Infof("provisioning virtual service for route %s...", route.Name)
				lbSpec := &model.LoadBalancerAVISpec{}
				if err := json.Unmarshal(lb.Spec, lbSpec); err != nil {
					panic(err)
				}

				if spec.ApplicationProfileID == "" {
					ap, err := findApplicationProfile(ac, defAVIVirtualServiceApplicationProfile)
					if err != nil {
						log.Errorf("failed to obtain application profile from load balancer %s due to: %v.", strconv.Quote(lb.Name), err)
						return nil
					}
					if ap == nil {
						log.Errorf("unable to find application profile %s.", defAVIVirtualServiceApplicationProfile)
						return fmt.Errorf("no such application profile: %s", defAVIVirtualServiceApplicationProfile)
					}
					spec.ApplicationProfileID = ap.UUID
				}
				if spec.NetworkProfileID == "" {
					var profileName string
					switch route.Protocol {
					case model.RouteProtocolUDP:
						profileName = "System-UDP-Fast-Path"
					default:
						// TCP
						profileName = "System-TCP-Fast-Path"
					}
					np, err := findNetworkProfile(ac, profileName)
					if err != nil {
						log.Errorf("failed to obtain default network profile %s from load balancer %s due to: %v.", profileName, strconv.Quote(lb.Name), err)
						return nil
					}
					if np == nil {
						log.Errorf("unable to find default network profile %s.", profileName)
						return fmt.Errorf("no such default network profile: %s", profileName)
					}
					spec.NetworkProfileID = np.UUID
				}

				vsName := "map-vs-" + standardizeString(route.Name) + "-" + route.OrganizationID.String()
				updates := make(map[string]interface{})
				if spec.VSID != "" {
					if _, err := updateVirtualService(ac, spec.VSID, vsName, int32(route.Port), spec, lbSpec.SEGroupID); err != nil {
						log.Errorf("failed to update virtual service of route %s due to: %v.", route.Name, err)
						return nil
					}
				} else {
					vs, err := createVirtualService(ac, vsName, int32(route.Port), spec, lbSpec.SEGroupID)
					if err != nil {
						log.Errorf("failed to create virtual service of route %s due to: %v.", route.Name, err)
						return nil
					}
					spec.VSID = vs.UUID
				}
				updates["spec"], err = json.Marshal(spec)
				if err != nil {
					panic(err)
				}
				updates["provisioned"] = model.RouteProvisionRoute
				updates["provision_phase"] = model.RouteProvisionFinish
				if err := tx.Model(route).Updates(updates).Error; err != nil {
					return err
				}
				log.Infof("provisioned virtual service of route %s. Pending next phase...", route.Name)
				next = time.Second

			case model.RouteProvisionFinish:
				updates := make(map[string]interface{})
				updates["provisioned"] = model.RouteProvisionFinish
				if err := tx.Model(route).Updates(updates).Error; err != nil {
					return err
				}
				log.Infof("route %s is successfully provisioned.", route.Name)
				finished = true
			}
			// additional load balancer
		}
		return nil
	}); err != nil {
		finished = false
		log.Warnf("error occurred: %v.", err)
		return time.Second
	}

	if finished {
		mapd.EventEmitterRoute.Next(&mapd.Event{
			Type: mapd.EventTypeRemove,
			OID:  id,
		})
	}
	return next
}

func findVSVIP(client *avi.AVI, name string) (ap *avi.VSVIP, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	list, err := client.ListVSVIPContext(ctx, &avi.ListVSVIPOptions{Name: name})
	if err != nil {
		return nil, err
	}
	if list.Count < 1 {
		return nil, nil
	}
	return list.Results[0], nil
}

func findApplicationProfile(client *avi.AVI, name string) (ap *avi.ApplicationProfile, err error) {
	list, err := client.ListApplicationProfile(&avi.ListApplicationProfileOptions{Name: name})
	if err != nil {
		return nil, err
	}
	if list.Count < 1 {
		return nil, nil
	}
	return list.Results[0], nil
}

func findNetworkProfile(client *avi.AVI, name string) (ap *avi.NetworkProfile, err error) {
	list, err := client.ListNetworkProfile(&avi.ListNetworkProfileOptions{Name: name})
	if err != nil {
		return nil, err
	}
	if list.Count < 1 {
		return nil, nil
	}
	return list.Results[0], nil
}

func findNetwork(client *avi.AVI, name string) (network *avi.Network, err error) {
	list, err := client.ListNetwork(&avi.ListNetworkOptions{Name: name})
	if err != nil {
		return nil, err
	}
	if list.Count < 1 {
		return nil, nil
	}
	return list.Results[0], nil
}

func createVirtualService(client *avi.AVI, name string, port int32, spec *model.RouteAVISpec, segroupRef string) (*avi.VirtualService, error) {
	obj := &avi.VirtualService{
		Name:                  name,
		Enabled:               true,
		Type:                  "VS_TYPE_NORMAL",
		ApplicationProfileRef: client.ApplicationProfileRef(spec.ApplicationProfileID).String(),
		NetworkProfileRef:     client.NetworkProfileRef(spec.NetworkProfileID).String(),
		PoolRef:               client.PoolRef(spec.PoolID).String(),
		Services: []*avi.Service{{
			Port:        port,
			EnableSSL:   false,
			EnableHTTP2: false,
		}},
		VSVipRef:   client.VSVIPRef(spec.VSVIPID).String(),
		SEGroupRef: client.SEGroupRef(segroupRef).String(),
	}
	return client.CreateVirtualService(obj)
}

func updateVirtualService(client *avi.AVI, uuid, name string, port int32, spec *model.RouteAVISpec, segroupRef string) (*avi.VirtualService, error) {
	obj := &avi.VirtualService{
		Name:                  name,
		Enabled:               true,
		Type:                  "VS_TYPE_NORMAL",
		ApplicationProfileRef: client.ApplicationProfileRef(spec.ApplicationProfileID).String(),
		NetworkProfileRef:     client.NetworkProfileRef(spec.NetworkProfileID).String(),
		PoolRef:               client.PoolRef(spec.PoolID).String(),
		Services: []*avi.Service{{
			Port:        port,
			EnableSSL:   false,
			EnableHTTP2: false,
		}},
		VSVipRef:   client.VSVIPRef(spec.VSVIPID).String(),
		SEGroupRef: client.SEGroupRef(segroupRef).String(),
		UUID:       uuid,
	}
	return client.UpdateVirtualService(obj)
}
