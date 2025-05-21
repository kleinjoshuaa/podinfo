package http

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/splitio/go-client/v6/splitio/client"
	"github.com/stefanprodan/podinfo/pkg/version"
)

// Info godoc
// @Summary Runtime information
// @Description returns the runtime information
// @Tags HTTP API
// @Accept json
// @Produce json
// @Success 200 {object} http.RuntimeResponse
// @Router /api/info [get]
func (s *Server) infoHandler(w http.ResponseWriter, r *http.Request) {
	_, span := s.tracer.Start(r.Context(), "infoHandler")
	defer span.End()

	// Generate a random user key that changes on every page refresh
	// Create a new random source with current time as seed
	randomSource := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNum := randomSource.Intn(1000) + 1 // Random number between 1-1000
	userKey := fmt.Sprintf("user-key-%d", randomNum)

	// Check the Split.io feature flag
	canaryEnabled := false
	// Check if the Split.io client has been initialized
	if s.splitInitialized {
		// Default to "off" treatment if the feature flag split doesn't exist
		splitClient, ok := s.splitClient.(*client.SplitClient)
		if ok {
			treatment := splitClient.Treatment(userKey, "podinfo_canary", nil)
			canaryEnabled = treatment == "on"
			s.logger.Sugar().Infof("Feature flag podinfo_canary treatment for %s: %s", userKey, treatment)
		} else {
			s.logger.Error("Failed to convert splitClient to *client.SplitClient")
		}
	}

	data := RuntimeResponse{
		Hostname:     s.config.Hostname,
		Version:      version.VERSION,
		Revision:     version.REVISION,
		Logo:         s.config.UILogo,
		Color:        s.config.UIColor,
		Message:      s.config.UIMessage,
		GOOS:         runtime.GOOS,
		GOARCH:       runtime.GOARCH,
		Runtime:      runtime.Version(),
		NumGoroutine: strconv.FormatInt(int64(runtime.NumGoroutine()), 10),
		NumCPU:       strconv.FormatInt(int64(runtime.NumCPU()), 10),
		CanaryEnabled: canaryEnabled,
		UserKey:      userKey,
	}

	s.JSONResponse(w, r, data)
}

type RuntimeResponse struct {
	Hostname     string `json:"hostname"`
	Version      string `json:"version"`
	Revision     string `json:"revision"`
	Color        string `json:"color"`
	Logo         string `json:"logo"`
	Message      string `json:"message"`
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
	Runtime      string `json:"runtime"`
	NumGoroutine string `json:"num_goroutine"`
	NumCPU       string `json:"num_cpu"`
	CanaryEnabled bool   `json:"canary_enabled"`
	UserKey      string `json:"user_key"`
}
