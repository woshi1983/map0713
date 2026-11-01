package submodule

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/labstack/echo/v4"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gorm.io/gorm"
)

const (
	defaultBrandNameZh    = "应用现代化平台"
	defaultBrandNameEn    = "Modern Application Platform"
	maxBrandingImageBytes = 256 * 1024
)

type brandingResponse struct {
	NameZh  string `json:"nameZh"`
	NameEn  string `json:"nameEn"`
	LogoURL string `json:"logoUrl"`
}

type brandingUpdate struct {
	NameZh      string  `json:"nameZh"`
	NameEn      string  `json:"nameEn"`
	LogoDataURL *string `json:"logoDataURL"`
	ResetLogo   bool    `json:"resetLogo"`
}

func validateBrandingName(value string) (string, error) {
	name := strings.TrimSpace(value)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 60 {
		return "", errors.New("brand name must contain 1 to 60 characters")
	}
	for _, char := range name {
		if unicode.IsControl(char) || char == '<' || char == '>' {
			return "", errors.New("brand name contains unsupported characters")
		}
	}
	return name, nil
}

func parseBrandingLogo(value string) ([]byte, string, error) {
	if len(value) == 0 || len(value) > maxBrandingImageBytes*2 {
		return nil, "", errors.New("logo must be at most 256 KiB")
	}
	var mime, prefix string
	for _, candidate := range []struct{ mime, prefix string }{
		{"image/png", "data:image/png;base64,"},
		{"image/jpeg", "data:image/jpeg;base64,"},
	} {
		if strings.HasPrefix(value, candidate.prefix) {
			mime, prefix = candidate.mime, candidate.prefix
			break
		}
	}
	if mime == "" {
		return nil, "", errors.New("logo must be PNG or JPEG")
	}
	data, err := base64.StdEncoding.Strict().DecodeString(strings.TrimPrefix(value, prefix))
	if err != nil || len(data) == 0 || len(data) > maxBrandingImageBytes {
		return nil, "", errors.New("invalid or oversized logo")
	}
	if http.DetectContentType(data) != mime {
		return nil, "", errors.New("logo format does not match its content")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || format != strings.TrimPrefix(mime, "image/") || config.Width < 1 || config.Height < 1 || config.Width > 2048 || config.Height > 1024 {
		return nil, "", errors.New("logo must be a valid image up to 2048 x 1024 pixels")
	}
	return data, mime, nil
}

func loadBranding() (*model.BrandingSetting, error) {
	setting := &model.BrandingSetting{ID: 1, NameZh: defaultBrandNameZh, NameEn: defaultBrandNameEn}
	err := mapd.Database().Where("id = ?", 1).Take(setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &model.BrandingSetting{ID: 1, NameZh: defaultBrandNameZh, NameEn: defaultBrandNameEn}, nil
	}
	return setting, err
}

func brandingView(setting *model.BrandingSetting) brandingResponse {
	view := brandingResponse{NameZh: setting.NameZh, NameEn: setting.NameEn}
	if len(setting.Logo) > 0 {
		view.LogoURL = "/api/branding/logo"
	}
	return view
}

func (m *API) getBranding(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	setting, err := loadBranding()
	if err != nil {
		return echo.ErrServiceUnavailable
	}
	return c.JSON(http.StatusOK, brandingView(setting))
}

func (m *API) getBrandingLogo(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	setting, err := loadBranding()
	if err != nil {
		return echo.ErrServiceUnavailable
	}
	if len(setting.Logo) == 0 {
		return echo.ErrNotFound
	}
	return c.Blob(http.StatusOK, setting.LogoMIME, setting.Logo)
}

func (m *API) putBranding(c echo.Context) error {
	sess, ok := auth.FromContext(c.Request().Context())
	if !ok {
		return echo.ErrUnauthorized
	}
	if !sess.Admin {
		return echo.ErrForbidden
	}
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxBrandingImageBytes*2+4096)
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	var input brandingUpdate
	if err := decoder.Decode(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid branding request")
	}
	if err := decoder.Decode(new(interface{})); err != io.EOF {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid branding request")
	}
	zh, err := validateBrandingName(input.NameZh)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	en, err := validateBrandingName(input.NameEn)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if input.ResetLogo && input.LogoDataURL != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "choose either logo upload or reset")
	}
	setting, err := loadBranding()
	if err != nil {
		return echo.ErrServiceUnavailable
	}
	setting.NameZh, setting.NameEn = zh, en
	if input.ResetLogo {
		setting.Logo, setting.LogoMIME = nil, ""
	}
	if input.LogoDataURL != nil {
		logo, mime, err := parseBrandingLogo(*input.LogoDataURL)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		setting.Logo, setting.LogoMIME = logo, mime
	}
	if err := mapd.Database().Save(setting).Error; err != nil {
		return echo.ErrServiceUnavailable
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	return c.JSON(http.StatusOK, brandingView(setting))
}
