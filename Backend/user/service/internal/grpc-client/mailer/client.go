package mailer

import (
	"auth/internal/config"
	"auth/internal/grpc-client/mailer/gen"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SenderClient struct {
	client   gen.SenderClient
	conn     *grpc.ClientConn
}

func NewSenderClient(cfg config.MailerGrpcClient) (*SenderClient, error) {
    conn, err := grpc.NewClient(
		cfg.Host+cfg.Port,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
    )
    if err != nil {
        return nil, err
    }
    return &SenderClient{
        client: gen.NewSenderClient(conn),
        conn:   conn,
    }, nil
}

func (sc *SenderClient) SendMsg(ctx context.Context, email, theme, msg string) (*gen.SendMailRes, error) {
	req := &gen.SendMailReq{
		Email:       email,
		Theme:       theme,
		Usermessage: msg,
	}
	return sc.client.SendMail(ctx, req)
}
