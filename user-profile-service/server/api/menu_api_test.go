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

const (
	testAPIMenu  = "/api/menu"
	testMenuID1  = "menu-1"
	testMenuID2  = "menu-2"
	testIconURL1 = "http://icon1.png"
	testIconURL2 = "http://icon2.png"
)

var testMenuCols = []string{"id", "index", "type", "title", "icon_url", "is_active"}

// ═══════════════════════════════════════════
// HTTP HandleGetAllMenus
// ═══════════════════════════════════════════

func TestHandleGetAllMenusSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnRows(sqlmock.NewRows(testMenuCols).
			AddRow(testMenuID1, 1, "REGULAR", "Transfer", testIconURL1, true).
			AddRow(testMenuID2, 2, "PREMIUM", "Investasi", testIconURL2, true))

	r := httptest.NewRequest(http.MethodGet, testAPIMenu, nil)
	w := httptest.NewRecorder()

	srv.HandleGetAllMenus(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetAllMenusEmpty(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnRows(sqlmock.NewRows(testMenuCols))

	r := httptest.NewRequest(http.MethodGet, testAPIMenu, nil)
	w := httptest.NewRecorder()

	srv.HandleGetAllMenus(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleGetAllMenusDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnError(fmt.Errorf(testDbError))

	r := httptest.NewRequest(http.MethodGet, testAPIMenu, nil)
	w := httptest.NewRecorder()

	srv.HandleGetAllMenus(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ═══════════════════════════════════════════
// HTTP HandleGetMenusByAccountType
// ═══════════════════════════════════════════

func TestHandleGetMenusByAccountTypeRegular(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WithArgs("REGULAR").
		WillReturnRows(sqlmock.NewRows(testMenuCols).
			AddRow(testMenuID1, 1, "REGULAR", "Transfer", testIconURL1, true))

	r := httptest.NewRequest(http.MethodGet, testAPIMenu+"/REGULAR", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountType", "REGULAR")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	srv.HandleGetMenusByAccountType(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetMenusByAccountTypeMissingType(t *testing.T) {
	srv, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodGet, testAPIMenu+"/", nil)
	w := httptest.NewRecorder()

	srv.HandleGetMenusByAccountType(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetMenusByAccountTypeDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnError(fmt.Errorf(testDbError))

	r := httptest.NewRequest(http.MethodGet, testAPIMenu+"/PREMIUM", nil)
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

func TestGetAllMenusGRPCSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnRows(sqlmock.NewRows(testMenuCols).
			AddRow(testMenuID1, 1, "REGULAR", "Transfer", testIconURL1, true).
			AddRow(testMenuID2, 2, "PREMIUM", "Investasi", testIconURL2, true))

	resp, err := srv.GetAllMenus(context.Background(), &pb.GetAllMenusRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Menus, 2)
	assert.Equal(t, "Transfer", resp.Menus[0].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllMenusGRPCDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnError(fmt.Errorf(testDbError))

	_, err := srv.GetAllMenus(context.Background(), &pb.GetAllMenusRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

// ═══════════════════════════════════════════
// gRPC: GetMenusByAccountType
// ═══════════════════════════════════════════

func TestGetMenusByAccountTypeGRPCSuccess(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WithArgs("REGULAR").
		WillReturnRows(sqlmock.NewRows(testMenuCols).
			AddRow(testMenuID1, 1, "REGULAR", "Transfer", testIconURL1, true))

	resp, err := srv.GetMenusByAccountType(context.Background(), &pb.GetMenusByAccountTypeRequest{AccountType: "REGULAR"})
	require.NoError(t, err)
	assert.Len(t, resp.Menus, 1)
	assert.Equal(t, "REGULAR", resp.Menus[0].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMenusByAccountTypeGRPCEmptyAccountType(t *testing.T) {
	srv, _ := newTestServer(t)
	_, err := srv.GetMenusByAccountType(context.Background(), &pb.GetMenusByAccountTypeRequest{})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGetMenusByAccountTypeGRPCDBError(t *testing.T) {
	srv, mock := newTestServer(t)
	mock.ExpectQuery(`FROM menu`).
		WillReturnError(fmt.Errorf(testDbError))

	_, err := srv.GetMenusByAccountType(context.Background(), &pb.GetMenusByAccountTypeRequest{AccountType: "PREMIUM"})
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}
