package todoist_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	todoist "github.com/smirnoffmg/go-todoist"
)

func ExampleNew() {
	api := todoist.New(os.Getenv("TODOIST_TOKEN"))
	_ = api
}

func ExampleClient_Tasks() {
	api := todoist.New(os.Getenv("TODOIST_TOKEN"))

	// The iterator follows pagination cursors automatically.
	for task, err := range api.Tasks(context.Background(), &todoist.GetTasksArgs{ProjectID: "220474322"}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(task.ID, task.Content)
	}
}

func ExampleClient_CreateTask() {
	api := todoist.New(os.Getenv("TODOIST_TOKEN"))

	task, err := api.CreateTask(context.Background(), todoist.CreateTaskArgs{
		Content:   "Buy milk",
		DueString: todoist.Ptr("tomorrow at 9am"),
		Priority:  todoist.Ptr(todoist.PriorityHigh),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(task.ID)
}

func ExampleError() {
	api := todoist.New(os.Getenv("TODOIST_TOKEN"))

	_, err := api.GetTask(context.Background(), "does-not-exist")
	var apiErr *todoist.Error
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode)
		if apiErr.Temporary() {
			// Retry after apiErr.RetryAfter.
		}
	}
}

func ExampleClient_Sync() {
	api := todoist.New(os.Getenv("TODOIST_TOKEN"))

	addProject := todoist.NewCommand("project_add", "proj-tmp", map[string]any{"name": "Launch"})
	addTask := todoist.NewCommand("item_add", "", map[string]any{
		"content":    "Draft announcement",
		"project_id": "proj-tmp", // resolved via temp_id mapping
	})

	resp, err := api.Sync(context.Background(), todoist.SyncRequest{
		ResourceTypes: []string{"projects", "items"},
		Commands:      []todoist.Command{addProject, addTask},
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := resp.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("created project", resp.TempIDMapping["proj-tmp"])
	// Persist resp.SyncToken for the next incremental sync.
}
