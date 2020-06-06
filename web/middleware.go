package web

import (
	"context"

	"net/http"
)

func (s *Server) FlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.Session.Get(r, "rjiosession")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Set some session values.
		flash, ok := session.Values["flash"]

		if ok {
			delete(session.Values, "flash")
		}

		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(r.Context(), "flash", flash)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
