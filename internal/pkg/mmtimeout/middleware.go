package mmtimeout

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mmrouter"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

/*
TimeoutMiddleware manages the timeout call of APIs, configurable for each API endpoint.

Example of usage of this middleware:
router.GET(

	fmt.Sprintf("/companies/:%s/", tc_auth.AUTH_COMPANY_URL_FIELD),
	tc_middleware.TimeoutMiddleware(time.Duration(1)*time.Second),
	... //other middlewares
	func(ctx *gin.Context) {
		... // your logic

You can set a duration in which duration format you are confortable with.
In case of timeout, a 408 error code is sent to the client.
*/
func TimeoutMiddleware(duration time.Duration) gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(duration),
		timeout.WithResponse(mmrouter.ReturnTimeOutError),
	)
}
