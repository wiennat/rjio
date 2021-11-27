package feed

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func ApiRouter() http.Handler {
	r := chi.NewRouter()
	r.Route("/feeds", func(r chi.Router) {
		r.Get("/", getFeedSourceListHandler)

		r.Route("/{sourceID}", func(r chi.Router) {
			r.Use(FeedSourceCtx)
			r.Get("/", getFeedSourceHandler)
			r.Get("/items", getFeedItemsHandler)
		})
	})
	return r
}

func getFeedSourceListHandler(w http.ResponseWriter, r *http.Request) {
	sources := DbListSource()

	var sourcePtrs []*Source
	for i := range sources {
		sourcePtrs = append(sourcePtrs, &sources[i])
	}

	if err := render.RenderList(w, r, NewFeedSourceListResponse(sourcePtrs)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func getFeedSourceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source := ctx.Value("source").(Source)

	if err := render.Render(w, r, NewFeedSourceResponse(&source)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

//
type FeedSourceResponse struct {
	*Source
}

func NewFeedSourceListResponse(sources []*Source) []render.Renderer {
	list := []render.Renderer{}
	for _, source := range sources {
		list = append(list, NewFeedSourceResponse(source))
	}

	return list
}

func NewFeedSourceResponse(source *Source) *FeedSourceResponse {
	resp := FeedSourceResponse{Source: source}
	return &resp
}

func (rd *FeedSourceResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
