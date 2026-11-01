// vks-license-issuer is an offline operator tool. Never deploy its private key with mapd.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/license"
)

func main() {
	generate := flag.Bool("generate-key", false, "generate a new Ed25519 key pair")
	uid := flag.String("supervisor-uid", "", "UID of Supervisor kube-system namespace")
	expires := flag.String("expires-at", "", "expiry in RFC3339 format")
	flag.Parse()
	if *generate {
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			fatal(err)
		}
		fmt.Fprintln(os.Stderr, "Keep the private key offline. Set MAP_VKS_LICENSE_PUBLIC_KEY on mapd.")
		fmt.Println("PUBLIC=" + base64.StdEncoding.EncodeToString(pub))
		fmt.Println("PRIVATE=" + base64.StdEncoding.EncodeToString(priv))
		return
	}
	private, err := base64.StdEncoding.DecodeString(os.Getenv("MAP_VKS_LICENSE_PRIVATE_KEY"))
	if err != nil || len(private) != ed25519.PrivateKeySize {
		fatal(fmt.Errorf("MAP_VKS_LICENSE_PRIVATE_KEY must be a base64 Ed25519 private key"))
	}
	if *uid == "" {
		fatal(fmt.Errorf("supervisor-uid is required"))
	}
	expiry, err := time.Parse(time.RFC3339, *expires)
	if err != nil || !time.Now().Before(expiry) {
		fatal(fmt.Errorf("expires-at must be a future RFC3339 timestamp"))
	}
	payload, err := json.Marshal(license.Payload{Version: 1, SupervisorUID: *uid, ExpiresAt: expiry.UTC().Format(time.RFC3339), Features: []string{"vks"}})
	if err != nil {
		fatal(err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	fmt.Println(encoded + "." + base64.RawURLEncoding.EncodeToString(ed25519.Sign(ed25519.PrivateKey(private), []byte(encoded))))
}

func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
