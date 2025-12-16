package closer

import (
	"app/internal/logger"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"go.uber.org/zap"
)

const shutdownTimeout = 5 * time.Second

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
}
type Closer struct {
	mu     sync.Mutex
	once   sync.Once
	done   chan struct{}
	funcs  []func(ctx context.Context) error
	logger Logger
}

var globalCloser = NewWithLogger(&logger.NoopLogger{})

func AddNamed(name string, f func(ctx context.Context) error) {
	globalCloser.AddNamed(name, f)
}

func Add(f func(ctx context.Context) error) {
	globalCloser.Add(f)
}

func SetLogger(logger Logger) {
	globalCloser.SetLogger(logger)
}

func CloseAll(ctx context.Context) error {
	return globalCloser.CloseAll(ctx)
}

func Configure(signals ...os.Signal) {
	globalCloser.handleSignal(signals...)
}

func New(signals ...os.Signal) *Closer {
	return NewWithLogger(&logger.NoopLogger{}, signals...)
}

func NewWithLogger(l Logger, signals ...os.Signal) *Closer {
	c := &Closer{
		done:   make(chan struct{}),
		logger: l,
	}

	if len(signals) > 0 {
		go c.handleSignal(signals...)
	}

	return c
}

func (c *Closer) handleSignal(signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	defer signal.Stop(ch)

	select {
	case sig := <-ch:
		c.logger.Info(context.Background(), "Получен сигнал, начинаем graceful shutdown",
			zap.String("signal", sig.String()),
		)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := c.CloseAll(shutdownCtx); err != nil {
			c.logger.Error(context.Background(), "❌ Ошибка при закрытии ресурсов", zap.Error(err))
		}

	case <-c.done:
	}
}

func (c *Closer) CloseAll(ctx context.Context) error {
	var result error
	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		if len(funcs) == 0 {
			c.logger.Info(ctx, "Нет функций для закрытия ")
			return
		}

		errCh := make(chan error, len(funcs))
		var wg sync.WaitGroup

		for i := len(funcs) - 1; i >= 0; i-- {
			f := funcs[i]

			wg.Add(1)
			go func(f func(ctx context.Context) error) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						errCh <- fmt.Errorf("panic: %v", r)
					}
				}()

				if err := f(ctx); err != nil {
					errCh <- err
				}
			}(f)
		}

		go func() {
			wg.Wait()
			close(errCh)
		}()

		for {
			select {
			case err, ok := <-errCh:
				if !ok {
					c.logger.Info(ctx, "✅ Все ресурсы успешно закрыты")
					return
				}
				c.logger.Error(ctx, "Ошибка при закрытии ресурса", zap.Error(err))
				if result == nil {
					result = err
				}
			case <-ctx.Done():
				c.logger.Error(ctx, "❌ Контекст отменён во время закрытия", zap.Error(ctx.Err()))
				if result == nil {
					result = ctx.Err()
				}
				return
			}
		}
	})
	return result
}

func (c *Closer) AddNamed(name string, f func(ctx context.Context) error) {
	c.Add(func(ctx context.Context) error {
		start := time.Now()
		c.logger.Info(ctx, "Закрываем"+name)
		err := f(ctx)
		d := time.Since(start)
		if err != nil {
			c.logger.Error(ctx, fmt.Sprintf("❌ Ошибка при закрытии %s (заняло %s): %v", name, d, err))
			return err
		}
		c.logger.Info(ctx, fmt.Sprintf("✅ %s закрыт (заняло %s)", name, d))
		return nil
	})
}

func (c *Closer) Add(f func(ctx context.Context) error) {
	c.mu.Lock()
	c.funcs = append(c.funcs, f)
	c.mu.Unlock()
}

func (c *Closer) SetLogger(logger Logger) {
	c.logger = logger
}
