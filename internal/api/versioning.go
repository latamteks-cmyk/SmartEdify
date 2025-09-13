package api

import (
    "fmt"
    "net/http"
    "regexp"
)

type APIVersion struct {
    Major int
    Minor int
    Patch int
}

func (v APIVersion) String() string {
    return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v APIVersion) IsCompatible(requested APIVersion) bool {
    // Backward compatibility: same major version
    return v.Major == requested.Major && v.Minor >= requested.Minor
}

type VersionHandler struct {
    supportedVersions map[string]http.Handler
    defaultVersion    string
    deprecatedVersions map[string]string // version -> deprecation notice
}

func NewVersionHandler() *VersionHandler {
    return &VersionHandler{
        supportedVersions: make(map[string]http.Handler),
        deprecatedVersions: make(map[string]string),
        defaultVersion: "v1.0.0",
    }
}

func (vh *VersionHandler) RegisterVersion(version string, handler http.Handler) {
    vh.supportedVersions[version] = handler
}

func (vh *VersionHandler) DeprecateVersion(version, notice string) {
    vh.deprecatedVersions[version] = notice
}

func (vh *VersionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    version := vh.extractVersion(r)
    
    // Check if version is deprecated
    if notice, deprecated := vh.deprecatedVersions[version]; deprecated {
        w.Header().Set("Deprecation", "true")
        w.Header().Set("Sunset", notice)
        w.Header().Set("Link", fmt.Sprintf("</api/%s>; rel=\"successor-version\"", vh.defaultVersion))
    }
    
    handler, exists := vh.supportedVersions[version]
    if !exists {
        http.Error(w, fmt.Sprintf("API version %s not supported", version), http.StatusNotFound)
        return
    }
    
    handler.ServeHTTP(w, r)
}

func (vh *VersionHandler) extractVersion(r *http.Request) string {
    // Try header first
    if version := r.Header.Get("API-Version"); version != "" {
        return version
    }
    
    // Try URL path
    re := regexp.MustCompile(`/api/(v\d+\.\d+\.\d+)/`)
    matches := re.FindStringSubmatch(r.URL.Path)
    if len(matches) > 1 {
        return matches[1]
    }
    
    // Default version
    return vh.defaultVersion
}