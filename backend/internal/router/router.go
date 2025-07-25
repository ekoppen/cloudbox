package router

import (
	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/handlers"
	"github.com/cloudbox/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Initialize creates and configures the router
func Initialize(cfg *config.Config, db *gorm.DB) *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.RateLimit(cfg))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	projectHandler := handlers.NewProjectHandler(db, cfg)
	dataHandler := handlers.NewDataHandler(db, cfg)
	storageHandler := handlers.NewStorageHandler(db, cfg)
	userHandler := handlers.NewUserHandler(db, cfg)
	messagingHandler := handlers.NewMessagingHandler(db, cfg)
	organizationHandler := handlers.NewOrganizationHandler(db, cfg)
	deploymentHandler := handlers.NewDeploymentHandler(db, cfg)
	backupHandler := handlers.NewBackupHandler(db, cfg)
	sshKeyHandler := handlers.NewSSHKeyHandler(db, cfg)
	webServerHandler := handlers.NewWebServerHandler(db, cfg)
	githubHandler := handlers.NewGitHubHandler(db, cfg)
	functionHandler := handlers.NewFunctionHandler(db, cfg)
	adminHandler := handlers.NewAdminHandler(db, cfg)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "1.0.0",
			"service": "cloudbox-api",
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.RequireAuth(cfg), authHandler.Logout)
			auth.GET("/me", middleware.RequireAuth(cfg), authHandler.GetProfile)
			auth.PUT("/me", middleware.RequireAuth(cfg), authHandler.UpdateProfile)
		}

		// GitHub API routes (global, requires authentication)
		github := api.Group("/github")
		github.Use(middleware.RequireAuth(cfg))
		github.Use(middleware.RequireAdminOrSuperAdmin())
		{
			github.POST("/validate", githubHandler.ValidateRepository)
			github.GET("/search", githubHandler.SearchRepositories)
			github.GET("/user/repositories", githubHandler.GetUserRepositories)
		}

		// Protected routes (requires authentication)
		protected := api.Group("/")
		protected.Use(middleware.RequireAuth(cfg))
		{
			// Organizations (accessible by admin and superadmin)
			organizations := protected.Group("/organizations")
			organizations.Use(middleware.RequireAdminOrSuperAdmin())
			{
				organizations.GET("", organizationHandler.ListOrganizations)
				organizations.POST("", organizationHandler.CreateOrganization)
				organizations.GET("/:id", organizationHandler.GetOrganization)
				organizations.PUT("/:id", organizationHandler.UpdateOrganization)
				organizations.DELETE("/:id", organizationHandler.DeleteOrganization)
				organizations.GET("/:id/projects", organizationHandler.GetOrganizationProjects)
			}

			// Projects (accessible by admin and superadmin)
			projects := protected.Group("/projects")
			projects.Use(middleware.RequireAdminOrSuperAdmin())
			{
				projects.GET("", projectHandler.ListProjects)
				projects.POST("", projectHandler.CreateProject)
				projects.GET("/:id", projectHandler.GetProject)
				projects.PUT("/:id", projectHandler.UpdateProject)
				projects.DELETE("/:id", projectHandler.DeleteProject)
				
				// Project-specific routes
				projects.GET("/:id/api-keys", projectHandler.ListAPIKeys)
				projects.POST("/:id/api-keys", projectHandler.CreateAPIKey)
				projects.DELETE("/:id/api-keys/:key_id", projectHandler.DeleteAPIKey)
				
				projects.GET("/:id/cors", projectHandler.GetCORSConfig)
				projects.PUT("/:id/cors", projectHandler.UpdateCORSConfig)
				
				// Project deployment infrastructure
				projects.GET("/:id/ssh-keys", sshKeyHandler.ListSSHKeys)
				projects.POST("/:id/ssh-keys", sshKeyHandler.CreateSSHKey)
				projects.GET("/:id/ssh-keys/:key_id", sshKeyHandler.GetSSHKey)
				projects.DELETE("/:id/ssh-keys/:key_id", sshKeyHandler.DeleteSSHKey)
				
				projects.GET("/:id/web-servers", webServerHandler.ListWebServers)
				projects.POST("/:id/web-servers", webServerHandler.CreateWebServer)
				projects.GET("/:id/web-servers/:server_id", webServerHandler.GetWebServer)
				projects.PUT("/:id/web-servers/:server_id", webServerHandler.UpdateWebServer)
				projects.DELETE("/:id/web-servers/:server_id", webServerHandler.DeleteWebServer)
				projects.POST("/:id/web-servers/:server_id/test", webServerHandler.TestConnection)
				
				projects.GET("/:id/github-repositories", githubHandler.ListGitHubRepositories)
				projects.POST("/:id/github-repositories", githubHandler.CreateGitHubRepository)
				projects.GET("/:id/github-repositories/:repo_id", githubHandler.GetGitHubRepository)
				projects.PUT("/:id/github-repositories/:repo_id", githubHandler.UpdateGitHubRepository)
				projects.DELETE("/:id/github-repositories/:repo_id", githubHandler.DeleteGitHubRepository)
				projects.POST("/:id/github-repositories/:repo_id/sync", githubHandler.SyncRepository)
				projects.GET("/:id/github-repositories/:repo_id/webhook", githubHandler.GetWebhookInfo)
				projects.GET("/:id/github-repositories/:repo_id/branches", githubHandler.GetRepositoryBranches)
				
				projects.GET("/:id/deployments", deploymentHandler.ListDeployments)
				projects.POST("/:id/deployments", deploymentHandler.CreateDeployment)
				projects.GET("/:id/deployments/:deployment_id", deploymentHandler.GetDeployment)
				projects.PUT("/:id/deployments/:deployment_id", deploymentHandler.UpdateDeployment)
				projects.DELETE("/:id/deployments/:deployment_id", deploymentHandler.DeleteDeployment)
				projects.POST("/:id/deployments/:deployment_id/deploy", deploymentHandler.Deploy)
				projects.GET("/:id/deployments/:deployment_id/logs", deploymentHandler.GetLogs)
				
				// Functions
				projects.GET("/:id/functions", functionHandler.ListFunctions)
				projects.POST("/:id/functions", functionHandler.CreateFunction)
				projects.GET("/:id/functions/:function_id", functionHandler.GetFunction)
				projects.PUT("/:id/functions/:function_id", functionHandler.UpdateFunction)
				projects.DELETE("/:id/functions/:function_id", functionHandler.DeleteFunction)
				projects.POST("/:id/functions/:function_id/deploy", functionHandler.DeployFunction)
				projects.POST("/:id/functions/:function_id/execute", functionHandler.ExecuteFunction)
				projects.GET("/:id/functions/:function_id/logs", functionHandler.GetFunctionLogs)
			}

			// Admin routes (accessible to authenticated users for demo)
			admin := protected.Group("/admin")
			{
				// Admin statistics endpoints
				admin.GET("/stats/system", adminHandler.GetSystemStats)
				admin.GET("/stats/user-growth", adminHandler.GetUserGrowth)
				admin.GET("/stats/project-activity", adminHandler.GetProjectActivity)
				admin.GET("/stats/function-executions", adminHandler.GetFunctionExecutions)
				admin.GET("/stats/deployment-stats", adminHandler.GetDeploymentStats)
				admin.GET("/stats/storage-stats", adminHandler.GetStorageStats)
				admin.GET("/stats/system-health", adminHandler.GetSystemHealth)
			}

			// Super Admin only routes
			superAdmin := protected.Group("/admin")
			superAdmin.Use(middleware.RequireSuperAdmin())
			{
				// Super admin can manage users
				superAdmin.GET("/users", authHandler.ListUsers)
				superAdmin.PUT("/users/:id/role", authHandler.UpdateUserRole)
				superAdmin.DELETE("/users/:id", authHandler.DeleteUser)
				
				// Super admin can see all projects (already handled in ListProjects)
				superAdmin.GET("/projects", projectHandler.ListProjects)
			}

			// Deployments
			deployments := protected.Group("/deployments")
			{
				deployments.GET("", deploymentHandler.ListDeployments)
				deployments.POST("", deploymentHandler.CreateDeployment)
				deployments.GET("/:id", deploymentHandler.GetDeployment)
				deployments.DELETE("/:id", deploymentHandler.DeleteDeployment)
				deployments.GET("/:id/logs", deploymentHandler.GetLogs)
			}

			// Backups
			backups := protected.Group("/backups")
			{
				backups.GET("", backupHandler.ListBackups)
				backups.POST("", backupHandler.CreateBackup)
				backups.GET("/:id", backupHandler.GetBackup)
				backups.DELETE("/:id", backupHandler.DeleteBackup)
				backups.POST("/:id/restore", backupHandler.RestoreBackup)
			}
		}
	}

	// Project API routes (project-specific namespaced APIs)
	projectAPI := r.Group("/p/:project_slug/api")
	projectAPI.Use(middleware.ProjectAuth(cfg, db))
	{
		// Collections management
		projectAPI.GET("/collections", dataHandler.ListCollections)
		projectAPI.POST("/collections", dataHandler.CreateCollection)
		projectAPI.GET("/collections/:collection", dataHandler.GetCollection)
		projectAPI.DELETE("/collections/:collection", dataHandler.DeleteCollection)
		
		// Documents management (legacy /data endpoints for compatibility)
		projectAPI.GET("/data/:collection", dataHandler.ListDocuments)
		projectAPI.POST("/data/:collection", dataHandler.CreateDocument)
		projectAPI.GET("/data/:collection/:id", dataHandler.GetDocument)
		projectAPI.PUT("/data/:collection/:id", dataHandler.UpdateDocument)
		projectAPI.DELETE("/data/:collection/:id", dataHandler.DeleteDocument)
		
		// Documents management (new /documents endpoints)
		projectAPI.GET("/documents/:collection", dataHandler.ListDocuments)
		projectAPI.POST("/documents/:collection", dataHandler.CreateDocument)
		projectAPI.GET("/documents/:collection/:id", dataHandler.GetDocument)
		projectAPI.PUT("/documents/:collection/:id", dataHandler.UpdateDocument)
		projectAPI.DELETE("/documents/:collection/:id", dataHandler.DeleteDocument)
		
		// Storage management
		projectAPI.GET("/storage/buckets", storageHandler.ListBuckets)
		projectAPI.POST("/storage/buckets", storageHandler.CreateBucket)
		projectAPI.GET("/storage/buckets/:bucket", storageHandler.GetBucket)
		projectAPI.DELETE("/storage/buckets/:bucket", storageHandler.DeleteBucket)
		
		// File management
		projectAPI.GET("/storage/:bucket/files", storageHandler.ListFiles)
		projectAPI.POST("/storage/:bucket/files", storageHandler.UploadFile)
		projectAPI.GET("/storage/:bucket/files/:file_id", storageHandler.GetFile)
		projectAPI.DELETE("/storage/:bucket/files/:file_id", storageHandler.DeleteFile)
		
		// User management
		projectAPI.GET("/users", userHandler.ListUsers)
		projectAPI.POST("/users", userHandler.CreateUser)
		projectAPI.GET("/users/:user_id", userHandler.GetUser)
		projectAPI.PUT("/users/:user_id", userHandler.UpdateUser)
		projectAPI.DELETE("/users/:user_id", userHandler.DeleteUser)
		
		// User authentication
		projectAPI.POST("/users/login", userHandler.LoginUser)
		projectAPI.POST("/users/logout", userHandler.LogoutUser)
		projectAPI.GET("/users/me", userHandler.GetCurrentUser)
		projectAPI.PUT("/users/:user_id/password", userHandler.ChangePassword)
		
		// Session management
		projectAPI.GET("/users/:user_id/sessions", userHandler.ListSessions)
		projectAPI.DELETE("/users/:user_id/sessions/:session_id", userHandler.RevokeSession)
		
		// Messaging API
		projectAPI.GET("/messaging/channels", messagingHandler.ListChannels)
		projectAPI.POST("/messaging/channels", messagingHandler.CreateChannel)
		projectAPI.GET("/messaging/channels/:channel_id", messagingHandler.GetChannel)
		projectAPI.PUT("/messaging/channels/:channel_id", messagingHandler.UpdateChannel)
		projectAPI.DELETE("/messaging/channels/:channel_id", messagingHandler.DeleteChannel)
		
		// Channel membership
		projectAPI.GET("/messaging/channels/:channel_id/members", messagingHandler.ListChannelMembers)
		projectAPI.POST("/messaging/channels/:channel_id/members", messagingHandler.JoinChannel)
		projectAPI.DELETE("/messaging/channels/:channel_id/members/:user_id", messagingHandler.LeaveChannel)
		
		// Messages
		projectAPI.GET("/messaging/channels/:channel_id/messages", messagingHandler.ListMessages)
		projectAPI.POST("/messaging/channels/:channel_id/messages", messagingHandler.SendMessage)
		projectAPI.GET("/messaging/channels/:channel_id/messages/:message_id", messagingHandler.GetMessage)
		projectAPI.PUT("/messaging/channels/:channel_id/messages/:message_id", messagingHandler.UpdateMessage)
		projectAPI.DELETE("/messaging/channels/:channel_id/messages/:message_id", messagingHandler.DeleteMessage)
		
		// Functions execution (public access for deployed functions)
		projectAPI.POST("/functions/:function_name", functionHandler.ExecuteFunctionByName)
		projectAPI.GET("/functions/:function_name", functionHandler.ExecuteFunctionByName)
		projectAPI.PUT("/functions/:function_name", functionHandler.ExecuteFunctionByName)
		projectAPI.DELETE("/functions/:function_name", functionHandler.ExecuteFunctionByName)
	}

	// Static file serving for deployments
	r.Static("/static", "./uploads/deployments")
	
	// Public file serving for storage buckets
	r.Static("/storage", "./uploads")

	return r
}