package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DABronskikh/ago-5/cmd/service/app/dto"
	"github.com/DABronskikh/ago-5/cmd/service/app/dto/mongoDTO"
	"github.com/DABronskikh/ago-5/cmd/service/app/middleware/authenticator"
	"github.com/DABronskikh/ago-5/cmd/service/app/middleware/authorizator"
	"github.com/DABronskikh/ago-5/cmd/service/app/middleware/cache"
	"github.com/DABronskikh/ago-5/cmd/service/app/middleware/identificator"
	"github.com/DABronskikh/ago-5/pkg/business"
	"github.com/DABronskikh/ago-5/pkg/security"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"time"
)

var (
	ErrUrlNotFound = errors.New("url not found")
)

const cacheTimeout = 50 * time.Millisecond

type Server struct {
	securitySvc *security.Service
	businessSvc *business.Service
	cache       *redis.Pool
	mux         chi.Router
}

func NewServer(securitySvc *security.Service, businessSvc *business.Service, mux chi.Router, cache *redis.Pool) *Server {
	return &Server{securitySvc: securitySvc, businessSvc: businessSvc, mux: mux, cache: cache}
}

func (s *Server) Init() error {
	identificatorMd := identificator.Identificator
	authenticatorMd := authenticator.Authenticator(identificator.Identifier, s.securitySvc.UserDetails)

	roleChecker := func(ctx context.Context, roles ...string) bool {
		userDetails, err := authenticator.Authentication(ctx)
		if err != nil {
			return false
		}
		return s.securitySvc.HasAnyRole(ctx, userDetails, roles...)
	}
	serviceRoleMd := authorizator.Authorizator(roleChecker, security.RoleService)

	// работа с redis
	cacheMd := cache.Cache(func(ctx context.Context, path string) (i []byte, err error) {
		userDetails, err := authenticator.Authentication(ctx)
		if err != nil {
			return nil, err
		}

		details, ok := userDetails.(*security.UserDetails)
		if !ok {
			return nil, err
		}

		path = fmt.Sprintf("users:%v:%v", details.ID, path)
		value, err := s.FromCache(ctx, path)
		if err != nil && errors.Is(err, redis.ErrNil) {
			return nil, cache.ErrNotInCache
		}
		log.Println("Cache value = ", value)
		return value, err
	}, func(ctx context.Context, path string, data []byte) error {
		return s.ToCache(ctx, path, data)
	})

	s.mux.With(middleware.Logger).Post("/api/users", s.register)
	s.mux.With(middleware.Logger).Post("/tokens", s.token)
	s.mux.With(middleware.Logger,
		identificatorMd,
		authenticatorMd).Get("/cards", s.getCards)

	s.mux.With(middleware.Logger,
		identificatorMd,
		authenticatorMd,
		cacheMd).Get("/operations", s.getOperations)

	s.mux.With(middleware.Logger,
		identificatorMd,
		authenticatorMd,
		serviceRoleMd).Post("/operations", s.createOperations)

	s.mux.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		prepareResponseErr(writer, ErrUrlNotFound, http.StatusNotFound)
	})

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) token(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	user := &dto.UserDTO{}
	err := decoder.Decode(user)
	if err != nil {
		prepareResponseErr(writer, err, http.StatusBadRequest)
		return
	}

	login := user.Login
	if login == "" {
		prepareResponseErr(writer, security.ErrRequiredLogin, http.StatusBadRequest)
		return
	}

	password := user.Password
	if password == "" {
		prepareResponseErr(writer, security.ErrRequiredPass, http.StatusBadRequest)
		return
	}

	token, err := s.securitySvc.Login(request.Context(), login, password)
	if err != nil {
		prepareResponseErr(writer, err, http.StatusBadRequest)
		return
	}

	data := &dto.TokenDTO{Token: token}
	prepareResponse(writer, data, http.StatusCreated)
	return
}

func (s *Server) register(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	user := &dto.UserDTO{}
	err := decoder.Decode(user)
	if err != nil {
		prepareResponseErr(writer, err, http.StatusBadRequest)
		return
	}

	login := user.Login
	if login == "" {
		prepareResponseErr(writer, security.ErrRequiredLogin, http.StatusBadRequest)
		return
	}

	password := user.Password
	if password == "" {
		prepareResponseErr(writer, security.ErrRequiredPass, http.StatusBadRequest)
		return
	}

	id, err := s.securitySvc.Register(request.Context(), login, password)

	if err == security.ErrUserDuplication {
		prepareResponseErr(writer, err, http.StatusInternalServerError)
		return
	}
	if err != nil {
		prepareResponseErr(writer, err, http.StatusBadRequest)
		return
	}

	data := &dto.IdDTO{Id: *id}
	prepareResponse(writer, data, http.StatusCreated)
	return
}

func (s *Server) getCards(writer http.ResponseWriter, request *http.Request) {
	userDetails, err := authenticator.Authentication(request.Context())
	if err != nil {
		prepareResponseErr(writer, err, http.StatusBadRequest)
		return
	}

	details, ok := userDetails.(*security.UserDetails)
	if !ok {
		return
	}

	cardDB := []*dto.CardDTO{}
	hasRole := false

	if s.securitySvc.HasAnyRole(request.Context(), userDetails, security.RoleAdmin) {
		cardDB, err = s.securitySvc.GetCardsAdmin(request.Context())
		hasRole = true
	}
	if s.securitySvc.HasAnyRole(request.Context(), userDetails, security.RoleUser) {
		cardDB, err = s.securitySvc.GetCardsUser(request.Context(), details.ID)
		hasRole = true
	}

	if err == security.ErrUserDuplication {
		prepareResponseErr(writer, err, http.StatusInternalServerError)
		return
	}

	if err != nil {
		prepareResponseErr(writer, err, http.StatusBadRequest)
		return
	}

	if hasRole {
		data := &dto.CardsDTO{Cards: cardDB}
		prepareResponse(writer, data, http.StatusOK)
		return
	}

	prepareResponseErr(writer, security.ErrNoAccess, http.StatusForbidden)
	return
}

func (s *Server) getOperations(writer http.ResponseWriter, request *http.Request) {
	userDetails, err := authenticator.Authentication(request.Context())
	if err != nil {
		prepareResponseErr(writer, err, http.StatusBadRequest)
		return
	}

	details, ok := userDetails.(*security.UserDetails)
	if !ok {
		return
	}

	operations := &[]mongoDTO.OperationDTO{}

	operations, err = s.securitySvc.GetOperations(request.Context(), details.ID)
	if err != nil {
		prepareResponseErr(writer, err, http.StatusBadRequest)
		return
	}

	prepareResponse(writer, operations, http.StatusOK)
	return
}

func (s *Server) createOperations(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	operations := &mongoDTO.OperationsDTO{}
	err := decoder.Decode(operations)
	if err != nil {
		prepareResponseErr(writer, err, http.StatusBadRequest)
		return
	}

	userId := operations.UserId
	if userId == 0 {
		prepareResponseErr(writer, mongoDTO.ErrRequiredOperationUserID, http.StatusBadRequest)
		return
	}

	data, err := s.securitySvc.SaveOperations(request.Context(), operations)

	if err != nil {
		prepareResponseErr(writer, err, http.StatusBadRequest)
		return
	}

	prepareResponse(writer, data, http.StatusOK)
	return
}

func prepareResponseErr(w http.ResponseWriter, err error, wHeader int) {
	log.Println(err)
	data := &dto.ErrDTO{Err: err.Error()}
	prepareResponse(w, data, wHeader)
}

func prepareResponse(w http.ResponseWriter, dto interface{}, wHeader int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(wHeader)

	respBody, err := json.Marshal(dto)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = w.Write(respBody)
	if err != nil {
		log.Println(err)
	}
}

func (s *Server) FromCache(ctx context.Context, key string) ([]byte, error) {
	conn, err := s.cache.GetContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	reply, err := redis.DoWithTimeout(conn, cacheTimeout, "GET", key)
	if err != nil {
		return nil, err
	}

	value, err := redis.Bytes(reply, err)
	if err != nil {
		return nil, err
	}

	return value, err
}

func (s *Server) ToCache(ctx context.Context, key string, value []byte) error {
	userDetails, err := authenticator.Authentication(ctx)
	if err != nil {
		return err
	}

	details, ok := userDetails.(*security.UserDetails)
	if !ok {
		return err
	}

	key = fmt.Sprintf("users:%v:%v", details.ID, key)

	conn, err := s.cache.GetContext(ctx)
	if err != nil {
		log.Print(err)
		return err
	}

	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	_, err = redis.DoWithTimeout(conn, cacheTimeout, "SET", key, value)
	if err != nil {
		log.Print(err)
	}
	return err
}
