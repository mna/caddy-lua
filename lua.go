package lua

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Lua{})
	httpcaddyfile.RegisterHandlerDirective("lua", parseCaddyfile)
}

// Lua implements an HTTP handler that runs a Lua script to handle the request.
type Lua struct {
	CallStackSize       int    `json:"call_stack_size,omitempty"`
	RegistrySize        int    `json:"registry_size,omitempty"`
	RegistryMaxSize     int    `json:"registry_max_size,omitempty"`
	RegistryGrowStep    int    `json:"registry_grow_step,omitempty"`
	MinimizeStackMemory bool   `json:"minimize_stack_memory,omitempty"`
	HandlerPath         string `json:"handler_path,omitempty"`

	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (Lua) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.lua",
		New: func() caddy.Module { return new(Lua) },
	}
}

// Provision implements caddy.Provisioner.
func (l *Lua) Provision(ctx caddy.Context) error {
	l.logger = ctx.Logger(l)
	return nil
}

// Validate implements caddy.Validator.
func (l *Lua) Validate() error {
	if l.HandlerPath == "" {
		return errors.New("the handler_path configuration option is required")
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (l Lua) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	L := lua.NewState()
	defer L.Close()
	if err := L.DoFile(l.HandlerPath); err != nil {
		return err
	}
	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (l *Lua) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	asInt := func() (int, error) {
		var s string
		if !d.AllArgs(&s) {
			return 0, d.ArgErr()
		}
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, err
		}
		return int(i), nil
	}

	for d.Next() {
		for d.NextBlock(0) {
			switch field := d.Val(); field {
			case "call_stack_size":
				i, err := asInt()
				if err != nil {
					return d.Errf("%s: %w", field, err)
				}
				l.CallStackSize = i

			case "registry_size":
				i, err := asInt()
				if err != nil {
					return d.Errf("%s: %w", field, err)
				}
				l.RegistrySize = i

			case "registry_max_size":
				i, err := asInt()
				if err != nil {
					return d.Errf("%s: %w", field, err)
				}
				l.RegistryMaxSize = i

			case "registry_grow_step":
				i, err := asInt()
				if err != nil {
					return d.Errf("%s: %w", field, err)
				}
				l.RegistryGrowStep = i

			case "minimize_stack_memory":
				if d.CountRemainingArgs() > 0 {
					return d.Errf("%s: %w", field, d.ArgErr())
				}

			case "handler_path":
				if !d.Args(&l.HandlerPath) {
					return d.Errf("%s: %w", field, d.ArgErr())
				}

			default:
				return d.Errf("%s: unknown configuration option", field)
			}
		}
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Lua.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var l Lua
	err := l.UnmarshalCaddyfile(h.Dispenser)
	return l, err
}

// interface guards
var (
	_ caddy.Provisioner           = (*Lua)(nil)
	_ caddyfile.Unmarshaler       = (*Lua)(nil)
	_ caddyhttp.MiddlewareHandler = (*Lua)(nil)
	_ caddy.Validator             = (*Lua)(nil)
)
