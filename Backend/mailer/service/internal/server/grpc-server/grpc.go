package grpcserver

import (
	"context"
	"log/slog"
	"mailer/internal/config"
	"mailer/internal/server/grpc-server/gen"
	"mailer/internal/utils"
	"net"

	"google.golang.org/grpc"
)

type SenderServerImpl struct {
	gen.UnimplementedSenderServer
	log    *slog.Logger
	sender utils.SenderInterface
}

func (ss *SenderServerImpl) SendMail(ctx context.Context, in *gen.SendMailReq) (*gen.SendMailRes, error) {
	if err := ss.sender.SendMsg(in.Email, in.Theme, in.Usermessage); err != nil {
		ss.log.Warn("failed to send message", "err", err)
		return nil, err
	}
	return &gen.SendMailRes{}, nil
}

func NewServer(log *slog.Logger, cfg *config.Config) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Error("Failed to listen", slog.Any("Error: ", err))
		return nil, err
	}
	s := grpc.NewServer()

	sender, err := utils.NewSender(cfg.Sender, log)
	if err != nil {
		log.Warn("sender is not initialized", "err", err)
		return nil, err
	}

	var srv gen.SenderServer = &SenderServerImpl{
		sender: sender,
		log: log,
	}

	gen.RegisterSenderServer(s, srv)

	go func() {
		log.Info("MailerService started", slog.String("addr: ", cfg.Addr))
		if err := s.Serve(lis); err != nil {
			log.Error("Failed to serve", slog.Any("Error: ", err))
		}
	}()

	return s, nil
}
