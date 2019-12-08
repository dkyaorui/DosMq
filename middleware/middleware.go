package middleware

import (
    "DosMq/modules"
    "DosMq/utils"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "net/http"
)

func CheckAuthKeyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        var authKey modules.AuthKey
        var requestResult utils.RequestResult
        var err error
        if err = c.ShouldBindBodyWith(&authKey, binding.JSON); err != nil {
            requestResult = utils.RequestResult{
                Code: http.StatusPartialContent,
                Data: err.Error(),
            }
            c.AbortWithStatusJSON(http.StatusNonAuthoritativeInfo, requestResult)
        } else if err = utils.CheckRegisterTopicKey(authKey.Value); err != nil {
            requestResult = utils.RequestResult{
                Code: http.StatusNonAuthoritativeInfo,
                Data: err.Error(),
            }
            c.AbortWithStatusJSON(http.StatusNonAuthoritativeInfo, requestResult)
        } else {
            c.Next()
        }
    }
}
