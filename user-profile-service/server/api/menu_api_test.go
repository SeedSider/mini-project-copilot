package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/bankease/user-profile-service/protogen/user-profile-service"
)

var testMenuCols = []string{"id", "index", "type", "title", "icon_url", "is_active"}

// ═══════════════════════════════════════════
// HTTP HandleGetAllMenus
// ═══════════════════════════════════════════

func TestHandleGetAllMenus_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnRows(sqlmock.NewRows(testMenuCols).
			AddRow("menu-1", 1, "REGULAR", "Transfer", "http://icon1.png", true).
			AddRow("menu-2", 2, "PREMIUM", "Investasi", "http://icon2.png", true))

	r := httptest.NewRequest(http.MethodGet, "/api/menu", nil)
	w := httptest.NewRecorder()

	srv.HandleGetAllMenus(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetAllMenus_Empty(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnRows(sqlmock.NewRows(testMenuCols))

	r := httptest.NewRequest(http.MethodGet, "/api/menu", nil)
	w := httptest.NewRecorder()

	srv.HandleGetAllMenus(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleGetAllMenus_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnError(fmt.Errorf("db error"))

	r := httptest.NewRequest(http.MethodGet, "/api/menu", nil)
	w := httptest.NewRecorder()

	srv.HandleGetAllMenus(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleGetMenusByAccountType
// ═══════════════════════════════════════════

func TestHandleGetMenusByAccountType_Regular(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WithArgs("REGULAR").
		WillReturnRows(sqlmock.NewRows(testMenuCols).
			AddRow("menu-1", 1, "REGULAR", "Transfer", "http://icon1.png", true))

	r := httptest.NewRequest(http.MethodGet, "/api/menu/REGULAR", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountType", "REGULAR")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	srv.HandleGetMenusByAccountType(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetMenusByAccountType_MissingType(t *testing.T) {
	srv, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodGet, "/api/menu/", nil)
	w := httptest.NewRecorder()

	srv.HandleGetMenusByAccountType(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetMenusByAccountType_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnError(fmt.Errorf("db error"))

	r := httptest.NewRequest(http.MethodGet, "/api/menu/PREMIUM", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountType", "PREMIUM")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	srv.HandleGetMenusByAccountType(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// gRPC: GetAllMenus
// ═══════════════════════════════════════════

func TestGetAllMenusGRPC_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnRows(sqlmock.NewRows(testMenuCols).
			AddRow("menu-1", 1, "REGULAR", "Transfer", "http://icon1.png", true).
			AddRow("menu-2", 2, "PREMIUM", "Investasi", "http://icon2.png", true))

	resp, err := srv.GetAllMenus(context.Background(), &pb.GetAllMenusRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Menus, 2)
	assert.Equal(t, "Transfer", resp.Menus[0].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllMenusGRPC_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.GetAllMenus(context.Background(), &pb.GetAllMenusRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// gRPC: GetMenusByAccountType
// ═══════════════════════════════════════════

func TestGetMenusByAccountTypeGRPC_Success(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WithArgs("REGULAR").
		WillReturnRows(sqlmock.NewRows(testMenuCols).
			AddRow("menu-1", 1, "REGULAR", "Transfer", "http://icon1.png", true))

	resp, err := srv.GetMenusByAccountType(context.Background(), &pb.GetMenusByAccountTypeRequest{AccountType: "REGULAR"})
	require.NoError(t, err)
	assert.Len(t, resp.Menus, 1)
	assert.Equal(t, "REGULAR", resp.Menus[0].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMenusByAccountTypeGRPC_EmptyAccountType(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.GetMenusByAccountType(context.Background(), &pb.GetMenusByAccountTypeRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGetMenusByAccountTypeGRPC_DBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnError(fmt.Errorf("db error"))

	_, err := srv.GetMenusByAccountType(context.Background(), &pb.GetMenusByAccountTypeRequest{AccountType: "PREMIUM"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}
