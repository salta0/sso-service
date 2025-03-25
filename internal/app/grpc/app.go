package grpcapp

import (
	"fmt"
	"log/slog"
	"net"
	authgrpc "sso/internal/grpc/auth"
	"strconv"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	port       int
	gRPCServer *grpc.Server
}

func New(log *slog.Logger, authService authgrpc.Auth, port int) *App {
	gRPCServer := grpc.NewServer()
	authgrpc.Register(gRPCServer, authService)

	return &App{log, port, gRPCServer}
}

func (app *App) MustRun() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func (app *App) Run() error {
	const op = "grpcapp.Run"

	log := app.log.With(slog.String("op", op), slog.Int("port", app.port))

	l, err := net.Listen("tcp", ":"+strconv.Itoa(app.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server is running", slog.String("addr", l.Addr().String()))
	if err := app.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (app *App) Stop() {
	const op = "frpcapp.Stop"

	app.log.With(slog.String("op", op), slog.Int("port", app.port)).Info("stopping gRPC server")

	app.gRPCServer.GracefulStop()
}
