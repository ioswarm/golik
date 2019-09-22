package golik

import (
	"fmt"
	"net/http"
	"strings"
	"encoding/json"
	"encoding/xml"
	"github.com/gorilla/mux"
)

type RouteContext struct {
	HttpClove CloveRef
	Request *http.Request
}

func (c *RouteContext) Params() map[string]string {
	return mux.Vars(c.Request)
}

// TODO append GetInt, GetLong and so on to RouteContext

type RouteHandler func(*RouteContext) Response

func NewRoute(path string, handler RouteHandler, methods ...string) *Route {
	result := &Route{
		path: path,
		handler: handler,
		methods: methods,
	}

	if len(result.methods) == 0 {
		result.methods = []string{"GET"}
	}

	return result
}

func Path(path string) *Route {
	return &Route{path: path, methods: []string{"GET"}}
}

func GET(path string) *Route {
	return Path(path)
}

func POST(path string) *Route {
	return &Route{path: path, methods: []string{"POST"}}
}

func PUT(path string) *Route {
	return &Route{path: path, methods: []string{"PUT"}}
}

func DELETE(path string) *Route {
	return &Route{path: path, methods: []string{"DELETE"}}
}

func PATCH(path string) *Route {
	return &Route{path: path, methods: []string{"PATCH"}}
}

// TODO OPTION and so on

type Route struct {
	path string
	handler RouteHandler
	methods []string
}

func (r *Route) String() string {
	var result string
	for i, m := range r.methods {
		if i > 0 {
			result += ", "
		}
		result += m
	}
	return strings.TrimLeft(result+" "+r.path, " ")
}

func (r *Route) Path(path string) *Route {
	r.path = path
	return r
}

func (r *Route) Handle(handler RouteHandler) *Route {
	r.handler = handler
	return r
}

func (r *Route) Method(methods ...string) *Route {
	r.methods = methods
	return r
}


/* Marshaller */

type ResponseBuild func(int, http.ResponseWriter)
type Response func(http.ResponseWriter)

func JSON(data interface{}) ResponseBuild {
	return func(status int, w http.ResponseWriter) {
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(data)
	}
}

func XML(data interface{}) ResponseBuild {
	return func(status int, w http.ResponseWriter) {
		w.Header().Add("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(status)
		xml.NewEncoder(w).Encode(data)
	}
}

/* Respond */

func buildResponse(status int, content interface{}) Response {
	return func (w http.ResponseWriter) {
		switch content.(type) {
		case string, int, int8, int16, int32, int64, bool, byte, float32, float64:
			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(status)
			fmt.Fprintf(w, "%v", content)
		// TODO handle error as InternalServerError as default maybe nil in OK and Created
		case Response:
			content.(Response)(w)
		case ResponseBuild:
			content.(ResponseBuild)(status, w)
		default:
			JSON(content)(status, w)
		}
	}
}

func OK                            (content interface{}) Response { return buildResponse(http.StatusOK                            , content) }
func Created                       (content interface{}) Response { return buildResponse(http.StatusCreated                       , content) }
func NoContent                     (content interface{}) Response { return buildResponse(http.StatusNoContent                     , content) }
func BadRequest                    (content interface{}) Response { return buildResponse(http.StatusBadRequest                    , content) }
func NotFound                      (content interface{}) Response { return buildResponse(http.StatusNotFound                      , content) }
func InternalServerError           (content interface{}) Response { return buildResponse(http.StatusInternalServerError           , content) }


func Continue                      (content interface{}) Response { return buildResponse(http.StatusContinue                      , content) }
func SwitchingProtocols            (content interface{}) Response { return buildResponse(http.StatusSwitchingProtocols            , content) }
func Processing                    (content interface{}) Response { return buildResponse(http.StatusProcessing                    , content) }
func Accepted                      (content interface{}) Response { return buildResponse(http.StatusAccepted                      , content) }
func NonAuthoritativeInfo          (content interface{}) Response { return buildResponse(http.StatusNonAuthoritativeInfo          , content) }
func ResetContent                  (content interface{}) Response { return buildResponse(http.StatusResetContent                  , content) }
func PartialContent                (content interface{}) Response { return buildResponse(http.StatusPartialContent                , content) }
func MultiStatus                   (content interface{}) Response { return buildResponse(http.StatusMultiStatus                   , content) }
func AlreadyReported               (content interface{}) Response { return buildResponse(http.StatusAlreadyReported               , content) }
func IMUsed                        (content interface{}) Response { return buildResponse(http.StatusIMUsed                        , content) }
func MultipleChoices               (content interface{}) Response { return buildResponse(http.StatusMultipleChoices               , content) }
func MovedPermanently              (content interface{}) Response { return buildResponse(http.StatusMovedPermanently              , content) }
func Found                         (content interface{}) Response { return buildResponse(http.StatusFound                         , content) }
func SeeOther                      (content interface{}) Response { return buildResponse(http.StatusSeeOther                      , content) }
func NotModified                   (content interface{}) Response { return buildResponse(http.StatusNotModified                   , content) }
func UseProxy                      (content interface{}) Response { return buildResponse(http.StatusUseProxy                      , content) }
func TemporaryRedirect             (content interface{}) Response { return buildResponse(http.StatusTemporaryRedirect             , content) }
func PermanentRedirect             (content interface{}) Response { return buildResponse(http.StatusPermanentRedirect             , content) }
func Unauthorized                  (content interface{}) Response { return buildResponse(http.StatusUnauthorized                  , content) }
func PaymentRequired               (content interface{}) Response { return buildResponse(http.StatusPaymentRequired               , content) }
func Forbidden                     (content interface{}) Response { return buildResponse(http.StatusForbidden                     , content) }
func MethodNotAllowed              (content interface{}) Response { return buildResponse(http.StatusMethodNotAllowed              , content) }
func NotAcceptable                 (content interface{}) Response { return buildResponse(http.StatusNotAcceptable                 , content) }
func ProxyAuthRequired             (content interface{}) Response { return buildResponse(http.StatusProxyAuthRequired             , content) }
func RequestTimeout                (content interface{}) Response { return buildResponse(http.StatusRequestTimeout                , content) }
func Conflict                      (content interface{}) Response { return buildResponse(http.StatusConflict                      , content) }
func Gone                          (content interface{}) Response { return buildResponse(http.StatusGone                          , content) }
func LengthRequired                (content interface{}) Response { return buildResponse(http.StatusLengthRequired                , content) }
func PreconditionFailed            (content interface{}) Response { return buildResponse(http.StatusPreconditionFailed            , content) }
func RequestEntityTooLarge         (content interface{}) Response { return buildResponse(http.StatusRequestEntityTooLarge         , content) }
func RequestURITooLong             (content interface{}) Response { return buildResponse(http.StatusRequestURITooLong             , content) }
func UnsupportedMediaType          (content interface{}) Response { return buildResponse(http.StatusUnsupportedMediaType          , content) }
func RequestedRangeNotSatisfiable  (content interface{}) Response { return buildResponse(http.StatusRequestedRangeNotSatisfiable  , content) }
func ExpectationFailed             (content interface{}) Response { return buildResponse(http.StatusExpectationFailed             , content) }
func Teapot                        (content interface{}) Response { return buildResponse(http.StatusTeapot                        , content) }
func MisdirectedRequest            (content interface{}) Response { return buildResponse(http.StatusMisdirectedRequest            , content) }
func UnprocessableEntity           (content interface{}) Response { return buildResponse(http.StatusUnprocessableEntity           , content) }
func Locked                        (content interface{}) Response { return buildResponse(http.StatusLocked                        , content) }
func FailedDependency              (content interface{}) Response { return buildResponse(http.StatusFailedDependency              , content) }
func TooEarly                      (content interface{}) Response { return buildResponse(http.StatusTooEarly                      , content) }
func UpgradeRequired               (content interface{}) Response { return buildResponse(http.StatusUpgradeRequired               , content) }
func PreconditionRequired          (content interface{}) Response { return buildResponse(http.StatusPreconditionRequired          , content) }
func TooManyRequests               (content interface{}) Response { return buildResponse(http.StatusTooManyRequests               , content) }
func RequestHeaderFieldsTooLarge   (content interface{}) Response { return buildResponse(http.StatusRequestHeaderFieldsTooLarge   , content) }
func UnavailableForLegalReasons    (content interface{}) Response { return buildResponse(http.StatusUnavailableForLegalReasons    , content) }
func NotImplemented                (content interface{}) Response { return buildResponse(http.StatusNotImplemented                , content) }
func BadGateway                    (content interface{}) Response { return buildResponse(http.StatusBadGateway                    , content) }
func ServiceUnavailable            (content interface{}) Response { return buildResponse(http.StatusServiceUnavailable            , content) }
func GatewayTimeout                (content interface{}) Response { return buildResponse(http.StatusGatewayTimeout                , content) }
func HTTPVersionNotSupported       (content interface{}) Response { return buildResponse(http.StatusHTTPVersionNotSupported       , content) }
func VariantAlsoNegotiates         (content interface{}) Response { return buildResponse(http.StatusVariantAlsoNegotiates         , content) }
func InsufficientStorage           (content interface{}) Response { return buildResponse(http.StatusInsufficientStorage           , content) }
func LoopDetected                  (content interface{}) Response { return buildResponse(http.StatusLoopDetected                  , content) }
func NotExtended                   (content interface{}) Response { return buildResponse(http.StatusNotExtended                   , content) }
func NetworkAuthenticationRequired (content interface{}) Response { return buildResponse(http.StatusNetworkAuthenticationRequired , content) }