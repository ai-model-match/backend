package mm_cors

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

/*
LocalhostOrigin represents an host can be used with CORS enabled
so the browser does not prevent API call from the client due to a different domain.
*/
const LocalhostOrigin = "http://localhost:5173"

/*
CorsMiddleware adds the needed headers in API response to enable CORS.
*/
func CorsMiddleware(allowOrigins []string) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT", "PATCH", "OPTIONS"},
		AllowHeaders:     append([]string{"content-type"}, cors.DefaultConfig().AllowHeaders...),
		AllowCredentials: true,
	})
}
