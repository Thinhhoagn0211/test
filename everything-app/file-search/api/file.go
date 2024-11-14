package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
	db "training/db/sqlc"
	"training/file-index/pb"
	"training/file-search/util"

	"github.com/gin-gonic/gin"
)

// @Summary Search files
// @Description Search files
// @Tags files
// @Accept  json
// @Produce  json
// @Param name query string false "File name"
// @Param extension query string false "File extension"
// @Param size_min query int false "Minimum file size"
// @Param size_max query int false "Maximum file size"
// @Param created_after query string false "Created after"
// @Param created_before query string false "Created before"
// @Param modified_after query string false "Modified after"
// @Param modified_before query string false "Modified before"
// @Param accessed_after query string false "Accessed after"
// @Param accessed_before query string false "Accessed before"
// @Param content query string false "Content"
// @Param offset query int false "Offset"
// @Param limit query int false "Limit"
// @Success 200 {object} newFileResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/files [get]
func (server *Server) getFileSearcher(ctx *gin.Context) {
	// Retrieve query parameters with defaults
	name := ctx.DefaultQuery("name", "")
	extension := ctx.DefaultQuery("extension", "")
	sizeMin, _ := strconv.Atoi(ctx.DefaultQuery("size_min", "0"))
	sizeMax, _ := strconv.Atoi(ctx.DefaultQuery("size_max", "1000000000000"))
	createdAfter := ctx.DefaultQuery("created_after", "950404073")
	createdBefore := ctx.DefaultQuery("created_before", "9999999999")
	modifiedAfter := ctx.DefaultQuery("modified_after", "950404073")
	modifiedBefore := ctx.DefaultQuery("modified_before", "9999999999")
	accessedAfter := ctx.DefaultQuery("accessed_after", "950404073")
	accessedBefore := ctx.DefaultQuery("accessed_before", "9999999999")

	// Parse timestamps as Unix times
	createdAfterTime, err := parseUnixTimestamp(createdAfter)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	createdBeforeTime, err := parseUnixTimestamp(createdBefore)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	modifiedAfterTime, err := parseUnixTimestamp(modifiedAfter)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	modifiedBeforeTime, err := parseUnixTimestamp(modifiedBefore)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	accessedAfterTime, err := parseUnixTimestamp(accessedAfter)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	accessedBeforeTime, err := parseUnixTimestamp(accessedBefore)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	fmt.Println(createdAfterTime, createdBeforeTime, modifiedAfterTime, modifiedBeforeTime, accessedAfterTime, accessedBeforeTime)
	content := ctx.DefaultQuery("content", "")
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "0"))
	// Prepare the database parameters
	arg := db.GetFilesParams{
		Column1:      util.NullableString(name),
		Extension:    extension,
		Size:         int64(sizeMin),
		Size_2:       int64(sizeMax),
		CreatedAt:    createdAfterTime,
		CreatedAt_2:  createdBeforeTime,
		ModifiedAt:   modifiedAfterTime,
		ModifiedAt_2: modifiedBeforeTime,
		AccessedAt:   accessedAfterTime,
		AccessedAt_2: accessedBeforeTime,
		Column11:     util.NullableString(content),
		Offset:       int32(offset),
		Column13:     int32(limit),
	}
	// Call the database to retrieve the files
	paths, err := server.store.GetFiles(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	req := &pb.CreateFileChecksumRequest{
		Filepath: paths,
	}

	var datas []newDataFile
	context := context.Background()
	responses, err := server.fileSearcherClient.GetCheckSumFiles(context, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	for key, value := range responses.Checksums {
		datas = append(datas, newDataFile{
			FilePath: key,
			CheckSum: value,
		})
	}

	response := newFileResponse{
		Status: http.StatusOK,
		Data:   datas,
		Meta: Metadata{
			Total:  len(paths),
			Offset: offset,
			Limit:  limit,
		},
	}
	// Return the search results
	ctx.JSON(http.StatusOK, response)
}

// Helper function to parse Unix timestamp
func parseUnixTimestamp(timestamp string) (time.Time, error) {
	unixTime, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(unixTime, 0), nil
}
