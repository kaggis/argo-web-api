package respond

import (
	"net/http"
	"strings"

	"github.com/ARGOeu/argo-web-api/utils/authentication"
	"github.com/ARGOeu/argo-web-api/utils/authorization"
	"github.com/ARGOeu/argo-web-api/utils/config"
	"github.com/gorilla/context"
)

// // WrapAll Wraps all wrap handlers. Note: Precedence is inversed
// func WrapAll(handler http.Handler, cfg config.Config) http.Handler {
//
// 	handler = WrapValidate(handler)
// 	handler = WrapAuthorize(handler)
// 	handler = WrapAuthenticate(handler, cfg)
//
// 	return handler
// }

func needsAPIAdmin(routeName string) bool {
	if strings.Split(routeName, ".")[0] == "tenants" {
		return true
	}

	return false
}

// WrapAuthenticate handle wrapper to apply authentication
func WrapAuthenticate(hfn http.Handler, cfg config.Config, routeName string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var errs []ErrorResponse

		// check if api admin authentication is needed (for tenants etc...)
		if needsAPIAdmin(routeName) {

			if (authentication.AuthenticateAdmin(r.Header, cfg)) == false {
				// Because not authenticated respond with error
				Error(w, r, ErrAuthen, cfg, errs)
				return
			}

			// admin api authenticated so continue serving
			context.Set(r, "authen", true)
			// Add admin restricted or not information -- used in get tenants

			// Check if admin is restricted
			if authentication.IsAdminRestricted(r.Header, cfg) {
				context.Set(r, "roles", []string{"super_admin_restricted"})
			} else if authentication.IsSuperAdminUI(r.Header, cfg) {
				context.Set(r, "roles", []string{"super_admin_ui"})
			} else {
				context.Set(r, "roles", []string{"super_admin"})
			}

			hfn.ServeHTTP(w, r)

		} else {

			// authenticate tenant user
			tenantConf, name, tErr := authentication.AuthenticateTenant(r.Header, cfg)
			// If tenant user not authenticated respond with  error
			if tErr != nil {
				Error(w, r, ErrAuthen, cfg, errs)
				return
			}

			context.Set(r, "roles", tenantConf.Roles)
			context.Set(r, "hbase_conf", cfg.Hbase)
			context.Set(r, "tenant_conf", tenantConf)
			context.Set(r, "tenant_name", name)
			context.Set(r, "authen", true)
			hfn.ServeHTTP(w, r)

		}

	})
}

// WrapAuthorize handle wrapper to apply authorization
func WrapAuthorize(hfn http.Handler, cfg config.Config, routeName string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var errs []ErrorResponse

		// log.Printf(" >> Authorization takes place here...")
		var roles []string

		roles = context.Get(r, "roles").([]string)

		if roles != nil {
			author := authorization.HasResourceRoles(cfg, routeName, roles)
			if author != false {
				hfn.ServeHTTP(w, r)
				return
			}
		}

		Error(w, r, ErrAuthor, cfg, errs)

	})
}

// WrapValidate handle wrapper to apply validation
func WrapValidate(hfn http.Handler, cfg config.Config, routeName string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var errs []ErrorResponse

		headers := r.Header
		queries := r.URL.Query()
		resource := strings.Split(routeName, ".")[0]

		// Validate Accept Header (globally unless OPTIONS verb is used)
		if r.Method != "OPTIONS" {
			err := ValidateAcceptHeader(headers.Get("Accept"))
			if err != (ErrorResponse{}) {
				Error(w, r, ErrValidHead, cfg, errs)
				return
			}
			if strings.Contains(resource, "status") {
				errs = ValidateStatusParams(queries)
				if len(errs) > 0 {
					Error(w, r, ErrValidQuery, cfg, errs)
					return
				}
			}
			if strings.Contains(resource, "results") {
				errs = ValidateResultsParams(queries)
				if len(errs) > 0 {
					Error(w, r, ErrValidQuery, cfg, errs)
					return
				}
			}
			if strings.Contains(resource, "metricResult") {
				errs = ValidateMetricParams(queries)
				if len(errs) > 0 {
					Error(w, r, ErrValidQuery, cfg, errs)
					return
				}
			}
		}

		hfn.ServeHTTP(w, r)
	})
}
