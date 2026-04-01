package grpcsrv

import (
	"context"
	"fmt"
	"net"
	"net/url"

	hdlr "github.com/nk87rus/go-musthave-shortener/internal/handler"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
	pb "github.com/nk87rus/go-musthave-shortener/internal/router/grpcsrv/proto"
	"github.com/nk87rus/go-musthave-shortener/internal/service/jwtproc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

//go:generate go run github.com/vektra/mockery/v2 --name=Handlers --inpackage --testonly
type Handlers interface {
	CreateShortURL(ctx context.Context, value string) (*url.URL, int, error)
	RestoreURL(ctx context.Context, id string) (string, bool, error)
	GetUsersURLs(ctx context.Context) ([]model.UsersURL, error)
}

type JWTProcessor interface {
	ParseJWT(jwtToken string) (string, error)
}

type Server struct {
	pb.UnimplementedShortenerServiceServer
	addr     string
	handlers Handlers
}

var (
	jwtProc JWTProcessor
)

func init() {
	jwtProc = jwtproc.New()
}

func New(address, baseAddress string, storage hdlr.Storage) (*Server, error) {
	baseURL, errURL := url.Parse(baseAddress)
	if errURL != nil {
		return nil, fmt.Errorf("не корректный base address: %w", errURL)
	}

	newSrv := Server{
		addr:     address,
		handlers: hdlr.InitHandlers(baseURL, storage),
	}
	return &newSrv, nil
}

func (gs *Server) Run(ctx context.Context) error {
	listen, err := net.Listen("tcp", gs.addr)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterShortenerServiceServer(s, gs)

	log.Info().Str("address", gs.addr).Msg("Запуск GRPC сервера")
	return s.Serve(listen)
}

// ShortenURL - создаёт сокращённый URL
func (gs *Server) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	var result pb.URLShortenResponse

	userID, err := getAuthData(ctx)
	if err != nil {
		log.Err(err)
		return &result, err
	}

	newURL, _, err := gs.handlers.CreateShortURL(context.WithValue(ctx, model.CtxUserID, userID), req.GetUrl())
	if err != nil {
		log.Err(err)
		return &result, err
	}

	result.SetResult(newURL.String())

	return &result, nil
}

// ExpandURL - восстанавливает исходный URL
func (gs *Server) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	var result pb.URLExpandResponse

	userID, err := getAuthData(ctx)
	if err != nil {
		log.Err(err)
		return &result, err
	}

	fullURL, isDeleted, err := gs.handlers.RestoreURL(context.WithValue(ctx, model.CtxUserID, userID), req.GetId())
	if err != nil {
		log.Err(err)
		return &result, err
	}

	if isDeleted {
		result.SetResult("URL deleted")
	} else {
		result.SetResult(fullURL)
	}

	return &result, nil
}

func (gs *Server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	var result pb.UserURLsResponse

	userID, err := getAuthData(ctx)
	if err != nil {
		log.Err(err)
		return &result, err
	}

	rawData, err := gs.handlers.GetUsersURLs(context.WithValue(ctx, model.CtxUserID, userID))
	if err != nil {
		log.Err(err)
		return &result, err
	}

	dataList := make([]*pb.URLData, 0, len(rawData))
	for _, u := range rawData {
		var item pb.URLData
		item.SetOriginalUrl(u.OrigURL)
		item.SetShortUrl(u.ShortURL)
		dataList = append(dataList, &item)
	}
	
	result.SetUrl(dataList)
	return &result, nil
}

func getAuthData(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("метаданные не найдены")
	}

	data := md.Get("authorization")
	if len(data) == 0 {
		return model.AnonymousUserID, nil
	}

	return jwtProc.ParseJWT(data[0])
}
