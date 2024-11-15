package api

import (
	"net/http"
	"strconv"
	"time"
	db "training/db/sqlc"
	"training/file-search/model"
	"training/file-search/util"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// @Summary Create a new user
// @Description Create a new user
// @Tags users
// @Accept  json
// @Produce  json
// @Param input body createUserRequest true "User info"
// @Success 200 {object} userResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users [post]
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.CreateUserParams{
		Email:        req.Email,
		Username:     req.Username,
		Password:     req.Password,
		PasswordHash: hashedPassword,
		Phone:        req.Phone,
		Fullname:     req.FullName,
		Avatar:       req.Avatar,
		State:        int64(model.UserStateActive),
		Role:         req.Role,
		CreatedAt:    time.Now(),
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	switch req.Role {
	case "admin":
		server.enforcer.AddPolicy("admin", "/*", "*")
	case "operator":
		server.enforcer.AddPolicy("operator", "/api/v1/files", "GET")
	}
	rsp := newUserResponse(user)
	ctx.JSON(http.StatusOK, rsp)
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:  user.Username,
		FullName:  user.Fullname,
		Role:      user.Role,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

// @Summary Get users
// @Description Get users
// @Tags users
// @Accept  json
// @Produce  json
// @Param search query string false "Search term"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param state query int false "State"
// @Param orderby query string false "Order by"
// @Param order query string false "Order"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users [get]
func (server *Server) getUsers(ctx *gin.Context) {
	searchTerm := ctx.DefaultQuery("search", "") // Get search term from query string
	state, _ := strconv.Atoi(ctx.DefaultQuery("state", "0"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))  // Default limit to 10
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0")) // Default offset to 0

	orderby := ctx.DefaultQuery("orderby", "id")
	order := ctx.DefaultQuery("order", "asc")
	var usersTotal []db.User
	if order == "asc" {
		// Fetch users with the search term and state filter
		users, err := server.store.GetUsersAsc(ctx, db.GetUsersAscParams{
			Column1: util.NullableString(searchTerm),
			State:   int64(state),
			Column3: util.NullableString(orderby),
			Limit:   int32(limit),
			Offset:  int32(offset),
		})
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		usersTotal = users
	} else {
		// Fetch users with the search term and state filter
		users, err := server.store.GetUsersDesc(ctx, db.GetUsersDescParams{
			Column1: util.NullableString(searchTerm),
			State:   int64(state),
			Column3: util.NullableString(orderby),
			Limit:   int32(limit),
			Offset:  int32(offset),
		})
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		usersTotal = users
	}
	rsp := Response{
		Status: http.StatusOK,
		Data:   usersTotal,
		Meta: Meta{
			Total:  len(usersTotal),
			Offset: offset,
			Limit:  limit,
		},
	}
	ctx.JSON(http.StatusOK, rsp)
}

// @Summary Get user by ID
// @Description Get user by ID
// @Tags users
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Success 200 {object} userResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users/{id} [get]
func (server *Server) getUserById(ctx *gin.Context) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	user, err := server.store.GetUserById(ctx, int32(idInt))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, user)
}

// @Summary Update user
// @Description Update user
// @Tags users
// @Accept  json
// @Produce  json
// @Param input body createUserRequest true "User info"
// @Success 200 {object} userResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users [patch]
func (server *Server) updateUser(ctx *gin.Context) {
	var req updateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateUserParams{
		Fullname:     req.FullName,
		Email:        req.Email,
		Phone:        req.Phone,
		Password:     req.Password,
		Avatar:       req.Avatar,
		Role:         req.Role,
		PasswordHash: hashedPassword,
		Username:     req.Username,
	}
	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, user)
}

// @Summary Delete user
// @Description Delete user
// @Tags users
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Success 200 {object} userResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users/{id} [delete]
func (server *Server) deleteUser(ctx *gin.Context) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	user, err := server.store.DeleteUser(ctx, db.DeleteUserParams{
		State: int64(model.UserStateDeleted),
		ID:    int32(idInt),
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, user)
}

// @Summary Login user
// @Description Login user
// @Tags users
// @Accept  json
// @Produce  json
// @Param input body loginUserRequest true "User info"
// @Success 200 {object} loginUserResponse
// @Failure 400 {object} ErrorResponse
// @Router /login [post]
func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	var rsp loginUserResponse

	// Bind the request body to the loginUserRequest struct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// Check if the user is "admin" with hardcoded credentials
	if req.Username == "admin" && req.Password == "admin" {
		server.role = "admin"
		// Create an access token for the admin user
		accessToken, _, err := server.tokenMaker.CreateToken(
			req.Username,
			server.config.AccessTokenDuration,
			server.role,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		// Prepare the response
		rsp = loginUserResponse{
			Status: 200,
			Error:  "",
			Errors: nil,
			Data: DataObject{
				AccessToken: accessToken,
			},
		}
	} else {
		// If the user is not "admin", check credentials in the database
		user, err := server.store.GetUserByUsername(ctx, req.Username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		// Check if the provided username matches the stored username
		if user.Username != req.Username {
			rsp = loginUserResponse{
				Status: 404,
				Error:  "invalid username or password",
				Errors: []string{"invalid username or password"},
				Data:   DataObject{},
			}
			ctx.JSON(http.StatusUnauthorized, rsp)
			return
		}
		server.role = user.Role
		// Validate the password
		if err := util.CheckPassword(req.Password, user.PasswordHash); err != nil {
			ctx.JSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// Prepare the response
		rsp = loginUserResponse{
			Status: 200,
			Error:  "",
			Errors: nil,
			Data:   DataObject{},
		}
	}

	ctx.JSON(http.StatusOK, rsp)
}
