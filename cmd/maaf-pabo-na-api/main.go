package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/muhammed-mamun/maaf-pabo-na-api/internal/config"
	"github.com/muhammed-mamun/maaf-pabo-na-api/internal/http/handlers/github"
	"github.com/muhammed-mamun/maaf-pabo-na-api/internal/utils/responses"
)

// type RoastParams struct {
// 	Name      string
// 	UserBio   string
// 	Followers int
// 	Repos     []string
// }

type requestPayload struct {
	Username string `json:"username"`
}

// func generateRoast(params RoastParams) (string, error) {
// 	prompt := fmt.Sprintf(
// 		"Create a humorously critical roast about the GitHub user '%s' based on their bio: '%s' and repositories: %v and followes %d",
// 		params.Name, params.UserBio, params.Repos, params.Followers,
// 	)

// 	apiKey := config.MustLoad().OpenAIAPI.APIKEY
// 	client := openai.NewClient(apiKey)

// 	resp, err := client.CreateCompletion(context.Background(), openai.CompletionRequest{
// 		Model:       "text-davinci-003",
// 		Prompt:      prompt,
// 		MaxTokens:   150,
// 		Temperature: 0.8,
// 	})

// 	if err != nil {
// 		return "", fmt.Errorf("failed to generate roast: %v", err)
// 	}

// 	if len(resp.Choices) == 0 {
// 		return "", fmt.Errorf("no choices return from OpenAI API")
// 	}

// 	// Return the response text
// 	return resp.Choices[0].Text, nil
// }

func main() {
	// Load config
	cfg := config.MustLoad()
	fmt.Println(cfg.GenAIAPI.APIKEY)
	// Database setup (if needed, make sure this part is properly implemented)

	// Setup router
	router := http.NewServeMux()

	//route definition
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API endpoint is working fine!"))
	})

	router.HandleFunc("/v1/api", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
			return
		}
		var req requestPayload

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.WriteJson(w, http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("invalid request payload %s", err),
			})
			return
		}
		username := req.Username

		if username == "" {
			responses.WriteJson(w, http.StatusBadRequest, map[string]string{
				"error": "username is required",
			})
			return
		}

		//github
		client, err := github.NewClient(r.Context())
		if err != nil {
			responses.WriteJson(w, http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("error creating github client: %s", err),
			})
			return
		}

		user, err := client.GetUser(r.Context(), username)
		if err != nil {
			responses.WriteJson(w, http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("error fetching github profile: %s", err),
			})
			return
		}

		repos, err := client.GetRepositories(r.Context(), username)
		if err != nil {
			responses.WriteJson(w, http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("error fetching github repositories: %s", err),
			})
			return
		}

		repoNames := []string{}

		for _, repo := range repos {
			repoNames = append(repoNames, fmt.Sprintf("%s - %s", repo.GetName(), repo.GetDescription()))
		}
		results := map[string]interface{}{
			"name":         user.GetName(),
			"userBio":      user.GetBio(),
			"followers":    user.GetFollowers(),
			"repositories": repoNames,
		}
		//prepare response data
		// roastparams := RoastParams{
		// 	Name:      user.GetName(),
		// 	UserBio:   user.GetBio(),
		// 	Followers: user.GetFollowers(),
		// 	Repos:     repoNames,
		// }

		// roast, err := generateRoast(roastparams)

		// if err != nil {
		// 	responses.WriteJson(w, http.StatusInternalServerError, map[string]string{
		// 		"error": fmt.Sprintf("error generating roast: %s", err),
		// 	})
		// 	return
		// }
		// response := map[string]interface{}{
		// 	"roast": roast,
		// }

		//send response as JSON
		if err := responses.WriteJson(w, http.StatusOK, results); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding JSON: %s", err), http.StatusInternalServerError)
		}

	})

	// Server setup
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	// Log server start
	slog.Info("server started", slog.String("address", cfg.Addr))

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server %s", err)
		}
	}()

	<-done

	slog.Info("shutting down the server")

	// Setting a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	}

	slog.Info("server shutdown successfully")
}
