package service

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/streadway/amqp"
)

func HandleProbeTask(ch *amqp.Channel) app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        task := CreateProbeTask()

        err := ProcessProbeTask(ch, task)
        if err != nil {
            c.String(500, "Failed to assign task")
        } else {
            c.String(200, "Task assigned successfully")
        }
    }
}
