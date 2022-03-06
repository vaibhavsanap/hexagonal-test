package paginate

import (
    "github.com/gin-gonic/gin"
    "github.com/spf13/cast"

    "go-hexagonal/config"
)

/**
 * @author Rancho
 * @date 2022/1/6
 */

func GetPage(c *gin.Context) int {
    page := cast.ToInt(c.Query("page"))
    if page <= 0 {
        return 1
    }

    return page
}

func GetPageSize(c *gin.Context) int {
    pageSize := cast.ToInt(c.Query("page_size"))
    if pageSize <= 0 {
        return config.Config.HTTPServer.DefaultPageSize
    }
    if pageSize > config.Config.HTTPServer.MaxPageSize {
        return config.Config.HTTPServer.MaxPageSize
    }

    return pageSize
}

func GetPageOffset(page, pageSize int) int {
    return (page - 1) * pageSize
}
