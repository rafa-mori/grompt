package cli

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	gl "github.com/rafa-mori/grompt/logger"
	"github.com/rafa-mori/grompt/utils"

	s "github.com/rafa-mori/grompt/internal/services/server"
	t "github.com/rafa-mori/grompt/internal/types"

	"github.com/spf13/cobra"
)

func ServerCmdList() []*cobra.Command {
	return []*cobra.Command{
		startServer(),
	}
}

func startServer() *cobra.Command {
	var debug bool

	var startCmd = &cobra.Command{
		Use: "start",
		Annotations: GetDescriptions([]string{
			"This command starts the Grompt server.",
			"This command initializes the Grompt server and starts waiting for help to build prompts.",
		}, false),
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				gl.SetDebug(true)
				gl.Log("debug", "Debug mode enabled")
			}

			cfg := &t.Config{
				Port:           utils.GetEnvOr("PORT", t.DefaultPort),
				ClaudeAPIKey:   utils.GetEnvOr("CLAUDE_API_KEY", ""),
				OllamaEndpoint: utils.GetEnvOr("OLLAMA_ENDPOINT", "http://localhost:11434"),
			}

			// Inicializar servidor
			server := s.NewServer(cfg)

			// Graceful shutdown
			go func() {
				c := make(chan os.Signal, 1)
				signal.Notify(c, os.Interrupt, syscall.SIGTERM)
				<-c
				fmt.Println("\n🛑 Encerrando servidor...")
				server.Shutdown()
				os.Exit(0)
			}()

			// Iniciar servidor
			if err := server.Start(); err != nil {
				log.Fatal("❌ Erro ao iniciar servidor:", err)
			}

			gl.Log("success", "Grompt server started successfully")
		},
	}

	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	return startCmd
}
