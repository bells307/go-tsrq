package tsrq_grpc

import (
	context "context"

	"github.com/bells307/go-tsrq/internal/queue"
	grpc "google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// GRPC-обработчик
type TSRQHandler struct {
	queue TSRQueue
}

type TSRQueue interface {
	// Добавить элемент в очередь
	Enqueue(ctx context.Context, id string, data any) error
	// Взять следующий элемент из очереди
	Dequeue(ctx context.Context) (*queue.DequeueResult, error)
	// Удалить элемент из очереди
	Remove(ctx context.Context, id string) error
	// Проверка элемента на существование в очереди
	Exists(ctx context.Context, id string) (bool, error)
	// Подсчет количества элементов
	Count(ctx context.Context) (int64, error)
}

func NewTSRQHandler(queue TSRQueue) *TSRQHandler {
	return &TSRQHandler{queue}
}

// Регистрация хэндлера GRPC
func (h *TSRQHandler) Register(s *grpc.Server) {
	RegisterTSRQServiceServer(s, h)
}

func (h *TSRQHandler) Enqueue(ctx context.Context, data *QueuedData) (*emptypb.Empty, error) {
	if err := h.queue.Enqueue(ctx, data.Id, data.Data); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *TSRQHandler) Dequeue(ctx context.Context, _ *emptypb.Empty) (*DequeueResponse, error) {
	res, err := h.queue.Dequeue(ctx)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return &DequeueResponse{MaybeData: &DequeueResponse_Null{}}, nil
	} else {
		data := QueuedData{Id: res.Id, Data: string(res.Data)}
		return &DequeueResponse{MaybeData: &DequeueResponse_Data{Data: &data}}, nil
	}
}

func (h *TSRQHandler) Remove(ctx context.Context, id *Id) (*emptypb.Empty, error) {
	if err := h.queue.Remove(ctx, id.Id); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *TSRQHandler) Exists(ctx context.Context, id *Id) (*ExistsResponse, error) {
	exists, err := h.queue.Exists(ctx, id.Id)
	if err != nil {
		return nil, err
	}

	return &ExistsResponse{Exists: exists}, nil
}

func (h *TSRQHandler) Count(ctx context.Context, _ *emptypb.Empty) (*CountResponse, error) {
	count, err := h.queue.Count(ctx)
	if err != nil {
		return nil, err
	}

	return &CountResponse{Count: count}, nil
}

func (h *TSRQHandler) mustEmbedUnimplementedTSRQServiceServer() {}
