package errors

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/sagacioushugo/oauth2"
)

func ErrorResponse(err error, ctx *context.Context, statusCode ...int) {
	var repsErr error
	var respDescription string
	var respStatusCode int

	if v, ok := oauth2.Descriptions[err]; ok {
		repsErr = err
		respDescription = v
		respStatusCode = oauth2.StatusCodes[err]
		if respStatusCode == 0 {
			respStatusCode = http.StatusForbidden
		}
	} else if v, ok := ErrValueMap[err]; ok {
		repsErr = err
		respDescription = v.ErrMsg
		respStatusCode = v.StatusCode
	} else {
		repsErr = ErrServerError
		respDescription = err.Error()
		respStatusCode = http.StatusBadRequest
	}
	resp := make(map[string]interface{})
	resp["error"] = repsErr.Error()
	resp["error_description"] = respDescription
	logs.Error(respDescription)

	if len(statusCode) > 0 {
		respStatusCode = statusCode[0]
	}
	ctx.ResponseWriter.WriteHeader(respStatusCode)
	if resErr := json.NewEncoder(ctx.ResponseWriter).Encode(resp); resErr != nil {
		logs.Error(resErr)
	}

}
