package extend

import (
	"github.com/labstack/echo"
)

type HandlerFunc func(c *Context) error

func Handle(h HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		c := &Context{ context: ctx }
		return h(c)
	}
}

func MiddlewareHandle(h HandlerFunc) echo.MiddlewareFunc {
	return func (next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			c := &Context{ context: ctx }
			if e := h(c); e != nil {
				return e
			}
			return next(ctx)
		}
	}
}