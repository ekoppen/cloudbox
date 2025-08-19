package router

import (
	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/handlers"
	"github.com/cloudbox/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Initialize creates and configures the CloudBox API router
// 
// Router implements API Architecture Standards:
// - Global Admin APIs:      /api/v1/{resource}         (JWT authentication)
// - Project Management:     /api/v1/projects/{id}/{resource} (JWT authentication)  
// - Project Data APIs:      /p/{project_slug}/api/{resource} (API Key authentication)
// - Public Project APIs:    /p/{project_slug}/api/{resource} (No authentication)
//
// Authentication Standards:
// - JWT Bearer Token: Admin/Management endpoints (/api/v1/*)
// - API Key (X-API-Key): Project data operations (/p/{slug}/api/*)
// - Single auth method per route group (no mixed authentication)
func Initialize(cfg *config.Config, db *gorm.DB) *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	// Global CORS middleware for admin and standard API requests
	r.Use(middleware.CORS(cfg))
	
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
	portfolioHandler := handlers.NewPortfolioHandler(db, cfg)
	templateHandler := handlers.NewTemplateHandler(db, cfg)
	templateDeploymentHandler := handlers.NewTemplateDeploymentHandler(db, cfg)
	compatibilityHandler := handlers.NewCompatibilityHandler(db, cfg)
	systemSettingsHandler := handlers.NewSystemSettingsHandler(db, cfg)
	projectGitHubHandler := handlers.NewProjectGitHubHandler(db, cfg)
	publicFileHandler := handlers.NewPublicFileHandler(db, cfg)
	pluginHandler := handlers.NewPluginHandler(db, cfg)
	scriptRunnerHandler := handlers.NewScriptRunnerHandler(db, cfg)

	// ===========================================
	// SYSTEM & HEALTH ENDPOINTS
	// ===========================================
	
	// Health check endpoint (no authentication required)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "1.0.0",
			"service": "cloudbox-api",
		})
	})

	// ===========================================
	// PUBLIC WEBHOOK ENDPOINTS
	// ===========================================
	
	// Public webhook endpoint (no authentication required)
	webhooks := r.Group("/api/v1/deploy")
	{
		webhooks.POST("/webhook/:repo_id", deploymentHandler.HandleWebhook)
	}

	// ===========================================
	// PUBLIC FILE SERVING ENDPOINTS
	// ===========================================
	
	// Public file access (no authentication required)
	public := r.Group("/public")
	{
		public.GET("/:project_slug/:bucket_name/*file_path", publicFileHandler.ServePublicFile)
	}

	// ===========================================
	// GLOBAL ADMIN API ROUTES (/api/v1/*)
	// JWT Bearer Token Authentication Required
	// ===========================================
	
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
			auth.PUT("/change-password", middleware.RequireAuth(cfg), authHandler.ChangePassword)
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

		// GitHub OAuth callback (public, no auth required)
		githubOAuth := api.Group("/github/oauth")
		{
			githubOAuth.GET("/callback", githubHandler.GitHubOAuthCallback)
		}
		
		// Project-specific GitHub OAuth callback (public, no auth required)
		projectGitHubOAuth := api.Group("/projects/:id/github/oauth")
		{
			projectGitHubOAuth.GET("/callback", projectGitHubHandler.HandleProjectOAuthCallback)
		}

		// Plugin API routes (public endpoints for active plugins only)
		plugins := api.Group("/plugins")
		{
			// Only allow public access to active plugins list
			plugins.GET("/active", pluginHandler.GetActivePlugins)
			
			// Script Runner Plugin Routes (require project-level authentication)
			scriptRunner := plugins.Group("/script-runner")
			scriptRunner.Use(middleware.RequireAuth(cfg))
			scriptRunner.Use(middleware.RequireAdminOrSuperAdmin())
			{
				scriptRunner.GET("/scripts/:projectId", scriptRunnerHandler.GetProjectScripts)
				scriptRunner.POST("/scripts/:projectId", scriptRunnerHandler.CreateScript)
				scriptRunner.PUT("/scripts/:projectId/:scriptId", scriptRunnerHandler.UpdateScript)
				scriptRunner.DELETE("/scripts/:projectId/:scriptId", scriptRunnerHandler.DeleteScript)
				scriptRunner.POST("/execute/:projectId/:scriptId", scriptRunnerHandler.ExecuteScript)
				scriptRunner.POST("/execute-raw/:projectId", scriptRunnerHandler.ExecuteRawSQL)
				scriptRunner.GET("/templates", scriptRunnerHandler.GetTemplates)
				scriptRunner.POST("/setup-project/:projectId/:templateName", scriptRunnerHandler.SetupProjectTemplate)
			}
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
				projects.GET("/:id/stats", projectHandler.GetProjectStats)
				
				// Project-specific routes
				projects.GET("/:id/api-keys", projectHandler.ListAPIKeys)
				projects.POST("/:id/api-keys", projectHandler.CreateAPIKey)
				projects.DELETE("/:id/api-keys/:key_id", projectHandler.DeleteAPIKey)
				
				projects.GET("/:id/cors", projectHandler.GetCORSConfig)
				projects.PUT("/:id/cors", projectHandler.UpdateCORSConfig)
				
				// Project notes
				projects.GET("/:id/notes", projectHandler.GetProjectNotes)
				projects.PUT("/:id/notes", projectHandler.UpdateProjectNotes)
				
				// Project deployment infrastructure
				projects.GET("/:id/ssh-keys", sshKeyHandler.ListSSHKeys)
				projects.POST("/:id/ssh-keys", sshKeyHandler.CreateSSHKey)
				projects.GET("/:id/ssh-keys/:key_id", sshKeyHandler.GetSSHKey)
				projects.PUT("/:id/ssh-keys/:key_id", sshKeyHandler.UpdateSSHKey)
				projects.DELETE("/:id/ssh-keys/:key_id", sshKeyHandler.DeleteSSHKey)
				
				projects.GET("/:id/web-servers", webServerHandler.ListWebServers)
				projects.POST("/:id/web-servers", webServerHandler.CreateWebServer)
				projects.GET("/:id/web-servers/:server_id", webServerHandler.GetWebServer)
				projects.PUT("/:id/web-servers/:server_id", webServerHandler.UpdateWebServer)
				projects.DELETE("/:id/web-servers/:server_id", webServerHandler.DeleteWebServer)
				projects.POST("/:id/web-servers/:server_id/test", webServerHandler.TestConnection)
				projects.POST("/:id/web-servers/:server_id/distribute-key", webServerHandler.DistributePublicKey)
				
				projects.GET("/:id/github-repositories", githubHandler.ListGitHubRepositories)
				projects.POST("/:id/github-repositories", githubHandler.CreateGitHubRepository)
				projects.GET("/:id/github-repositories/:repo_id", githubHandler.GetGitHubRepository)
				projects.PUT("/:id/github-repositories/:repo_id", githubHandler.UpdateGitHubRepository)
				projects.DELETE("/:id/github-repositories/:repo_id", githubHandler.DeleteGitHubRepository)
				projects.POST("/:id/github-repositories/:repo_id/sync", githubHandler.SyncRepository)
				projects.GET("/:id/github-repositories/:repo_id/webhook", githubHandler.GetWebhookInfo)
				projects.GET("/:id/github-repositories/:repo_id/branches", githubHandler.GetRepositoryBranches)
				
				// Repository analysis endpoints
				projects.POST("/:id/github-repositories/analyze", githubHandler.AnalyzeRepository)
				projects.GET("/:id/github-repositories/:repo_id/analysis", githubHandler.GetRepositoryAnalysis)
				projects.POST("/:id/github-repositories/:repo_id/analyze", githubHandler.AnalyzeAndSaveRepository)
				projects.POST("/:id/github-repositories/:repo_id/reanalyze", githubHandler.ReAnalyzeRepository)
				
				// GitHub authentication endpoints
				projects.PUT("/:id/github-repositories/:repo_id/token", githubHandler.UpdateRepositoryToken)
				projects.GET("/:id/github-repositories/:repo_id/test-access", githubHandler.TestRepositoryAccess)
				
				// CIP compliance check endpoint
				projects.POST("/:id/github-repositories/:repo_id/cip-check", githubHandler.CheckCIPCompliance)
				
				projects.POST("/:id/github-repositories/:repo_id/deploy-pending", deploymentHandler.DeployPendingUpdate)
				
				projects.GET("/:id/deployments", deploymentHandler.ListDeployments)
				projects.POST("/:id/deployments", deploymentHandler.CreateDeployment)
				projects.GET("/:id/deployments/:deployment_id", deploymentHandler.GetDeployment)
				projects.PUT("/:id/deployments/:deployment_id", deploymentHandler.UpdateDeployment)
				projects.DELETE("/:id/deployments/:deployment_id", deploymentHandler.DeleteDeployment)
				projects.POST("/:id/deployments/:deployment_id/deploy", deploymentHandler.ExecuteCIPDeployment)
				projects.GET("/:id/deployments/:deployment_id/logs", deploymentHandler.GetLogs)
				projects.GET("/:id/deployments/:deployment_id/status", deploymentHandler.GetStatus)
				
				// Port availability checking
				projects.POST("/:id/deployments/check-ports", deploymentHandler.CheckPortAvailability)
				
				// Functions
				projects.GET("/:id/functions", functionHandler.ListFunctions)
				projects.POST("/:id/functions", functionHandler.CreateFunction)
				projects.GET("/:id/functions/:function_id", functionHandler.GetFunction)
				projects.PUT("/:id/functions/:function_id", functionHandler.UpdateFunction)
				projects.DELETE("/:id/functions/:function_id", functionHandler.DeleteFunction)
				projects.POST("/:id/functions/:function_id/deploy", functionHandler.DeployFunction)
				projects.POST("/:id/functions/:function_id/execute", functionHandler.ExecuteFunction)
				projects.GET("/:id/functions/:function_id/logs", functionHandler.GetFunctionLogs)
				
				// Project GitHub configuration
				projects.GET("/:id/github/config", projectGitHubHandler.GetProjectGitHubConfig)
				projects.PUT("/:id/github/config", projectGitHubHandler.UpdateProjectGitHubConfig)
				projects.POST("/:id/github/config/test", projectGitHubHandler.TestProjectGitHubConfig)
				projects.GET("/:id/github/instructions", projectGitHubHandler.GetProjectGitHubInstructions)
				
				// Admin Storage management endpoints
				projects.GET("/:id/storage/buckets", storageHandler.AdminListBuckets)
				projects.POST("/:id/storage/buckets", storageHandler.AdminCreateBucket)
				projects.GET("/:id/storage/buckets/:bucket", storageHandler.AdminGetBucket)
				projects.PUT("/:id/storage/buckets/:bucket", storageHandler.AdminUpdateBucket)
				projects.DELETE("/:id/storage/buckets/:bucket", storageHandler.AdminDeleteBucket)
				
				// Admin File management endpoints
				projects.GET("/:id/storage/buckets/:bucket/files", storageHandler.AdminListFiles)
				projects.POST("/:id/storage/buckets/:bucket/files", storageHandler.AdminUploadFile)
				projects.GET("/:id/storage/buckets/:bucket/files/:file_id", storageHandler.AdminGetFile)
				projects.PUT("/:id/storage/buckets/:bucket/files/:file_id/move", storageHandler.AdminMoveFile)
				projects.DELETE("/:id/storage/buckets/:bucket/files/:file_id", storageHandler.AdminDeleteFile)
				
				// Admin Folder management endpoints
				projects.GET("/:id/storage/buckets/:bucket/folders", storageHandler.AdminListFolders)
				projects.POST("/:id/storage/buckets/:bucket/folders", storageHandler.AdminCreateFolder)
				projects.DELETE("/:id/storage/buckets/:bucket/folders", storageHandler.AdminDeleteFolder)
				
				// Admin Collections management endpoints
				projects.GET("/:id/collections", dataHandler.AdminListCollections)
				projects.POST("/:id/collections", dataHandler.AdminCreateCollection)
				projects.GET("/:id/collections/:collection", dataHandler.AdminGetCollection)
				projects.DELETE("/:id/collections/:collection", dataHandler.AdminDeleteCollection)
				
				// Admin Documents management endpoints
				projects.GET("/:id/collections/:collection/documents", dataHandler.AdminListDocuments)
				projects.POST("/:id/collections/:collection/documents", dataHandler.AdminCreateDocument)
				projects.GET("/:id/collections/:collection/documents/:document_id", dataHandler.AdminGetDocument)
				projects.PUT("/:id/collections/:collection/documents/:document_id", dataHandler.AdminUpdateDocument)
				projects.DELETE("/:id/collections/:collection/documents/:document_id", dataHandler.AdminDeleteDocument)
				
				// Admin bucket visibility management endpoints
				projects.PUT("/:id/storage/buckets/:bucket/visibility", storageHandler.AdminSetBucketVisibility)
				projects.GET("/:id/storage/buckets/:bucket/files/:file_id/public-url", storageHandler.AdminGetFilePublicURL)
				projects.GET("/:id/storage/public-buckets", storageHandler.AdminListPublicBuckets)
				
				// Messaging API (simplified - only endpoints that frontend uses)
				projects.GET("/:id/messaging/messages", messagingHandler.ListAllMessages)
				projects.GET("/:id/messaging/templates", messagingHandler.ListTemplates)
				projects.GET("/:id/messaging/stats", messagingHandler.GetMessagingStats)
				projects.GET("/:id/messaging/channels", messagingHandler.ListChannels)
				projects.DELETE("/:id/messaging/messages/:message_id", messagingHandler.DeleteMessage)
				
				// Admin User management endpoints for projects
				projects.GET("/:id/auth/users", userHandler.AdminListUsers)
				projects.POST("/:id/auth/users", userHandler.AdminCreateUser)
				projects.GET("/:id/auth/users/:user_id", userHandler.AdminGetUser)
				projects.PUT("/:id/auth/users/:user_id", userHandler.AdminUpdateUser)
				projects.DELETE("/:id/auth/users/:user_id", userHandler.AdminDeleteUser)
				projects.GET("/:id/auth/settings", userHandler.AdminGetAuthSettings)
				projects.PUT("/:id/auth/settings", userHandler.AdminUpdateAuthSettings)
				
				// Project-level plugin management routes
				projects.GET("/:id/plugins/available", pluginHandler.GetAvailablePlugins)
				projects.GET("/:id/plugins/installed", pluginHandler.GetInstalledPlugins)
				projects.POST("/:id/plugins/:plugin_name/install", pluginHandler.InstallPluginToProject)
				projects.POST("/:id/plugins/:plugin_name/enable", pluginHandler.EnablePluginForProject)
				projects.POST("/:id/plugins/:plugin_name/disable", pluginHandler.DisablePluginForProject)
				projects.DELETE("/:id/plugins/:plugin_name", pluginHandler.UninstallPluginFromProject)
				projects.PUT("/:id/plugins/:plugin_name/config", pluginHandler.UpdatePluginConfigForProject)
				projects.GET("/:id/plugins/:plugin_name/status", pluginHandler.GetPluginStatusForProject)
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
				
				// Admin system endpoints
				admin.GET("/system/info", adminHandler.GetSystemInfo)
				admin.POST("/system/restart", adminHandler.RestartSystem)
				admin.POST("/system/clear-cache", adminHandler.ClearCache)
				admin.POST("/system/backup", adminHandler.CreateBackup)
				
				// System settings endpoints (configurable settings)
				admin.GET("/system/settings", systemSettingsHandler.GetSystemSettings)
				admin.PUT("/system/settings/:key", systemSettingsHandler.UpdateSystemSetting)
				admin.GET("/system/github-instructions", systemSettingsHandler.GetGitHubInstructions)
				admin.POST("/system/test-github-oauth", systemSettingsHandler.TestGitHubOAuth)

				// Plugin management endpoints (admin/superadmin only with enhanced security)
				admin.GET("/plugins", pluginHandler.GetAllPlugins)
				admin.POST("/plugins/install", pluginHandler.InstallPlugin)
				admin.POST("/plugins/:pluginName/enable", pluginHandler.EnablePlugin)
				admin.POST("/plugins/:pluginName/disable", pluginHandler.DisablePlugin)
				admin.DELETE("/plugins/:pluginName", pluginHandler.UninstallPlugin)
				admin.POST("/plugins/reload", pluginHandler.ReloadPlugins)
				admin.POST("/plugins/:pluginName/reload", pluginHandler.HotReloadPlugin)
				admin.GET("/plugins/repositories", pluginHandler.GetApprovedRepositories)
				admin.GET("/plugins/audit-logs", pluginHandler.GetAuditLogs)
				
				// Plugin marketplace endpoints
				admin.GET("/plugins/marketplace", pluginHandler.GetMarketplace)
				admin.POST("/plugins/marketplace/add", pluginHandler.AddPluginToMarketplace) // Changed path to avoid conflicts
				admin.GET("/plugins/marketplace/search", pluginHandler.SearchMarketplace)
				admin.GET("/plugins/marketplace/:pluginName", pluginHandler.GetPluginDetails)
				admin.POST("/plugins/marketplace/install", pluginHandler.InstallFromMarketplace)
				
				// Plugin health and configuration endpoints
				admin.GET("/plugins/health", pluginHandler.GetPluginHealth)
				admin.PUT("/plugins/:pluginName/config", pluginHandler.UpdatePluginConfig)

				// Basic admin project endpoint (accessible to JWT authenticated users)
				admin.GET("/projects/:id", projectHandler.GetProject)
				
				// Security monitoring endpoints (admin only)
				admin.GET("/security/plugin-activity", pluginHandler.GetAuditLogs)
			}

			// Super Admin only routes
			superAdmin := protected.Group("/admin")
			superAdmin.Use(middleware.RequireSuperAdmin())
			{
				// Super admin can manage users
				superAdmin.GET("/users", authHandler.ListUsers)
				superAdmin.POST("/users", authHandler.CreateUser)
				superAdmin.PUT("/users/:id", authHandler.UpdateUser)
				superAdmin.PUT("/users/:id/role", authHandler.UpdateUserRole)
				superAdmin.DELETE("/users/:id", authHandler.DeleteUser)
				
				// Organization admin management
				superAdmin.GET("/organization-admins", authHandler.ListOrganizationAdmins)
				superAdmin.POST("/organization-admins", authHandler.AssignOrganizationAdmin)
				superAdmin.DELETE("/organization-admins/:user_id/:org_id", authHandler.RevokeOrganizationAdmin)
				
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
		
		// Note: Collection management moved to project data API routes
		// Admin can access collections via /p/{project_slug}/api/collections with proper auth
	}

	// ===========================================
	// PROJECT DATA API ROUTES (/p/{slug}/api/*)
	// API Key (X-API-Key) Authentication Required
	// ===========================================
	
	// Protected project routes (API key authentication required)
	projectAPI := r.Group("/p/:project_slug/api")
	projectAPI.Use(middleware.ProjectAuthOrJWT(cfg, db)) // API key authentication
	projectAPI.Use(middleware.ProjectCORS(cfg, db))
	{
		// Collections management
		projectAPI.GET("/collections", dataHandler.ListCollections)
		projectAPI.POST("/collections", dataHandler.CreateCollection)
		projectAPI.GET("/collections/:collection", dataHandler.GetCollection)
		projectAPI.DELETE("/collections/:collection", dataHandler.DeleteCollection)
		
		// Documents management (standardized endpoints)
		projectAPI.GET("/data/:collection", dataHandler.ListDocuments)
		projectAPI.POST("/data/:collection", dataHandler.CreateDocument)
		projectAPI.GET("/data/:collection/:id", dataHandler.GetDocument)
		projectAPI.PUT("/data/:collection/:id", dataHandler.UpdateDocument)
		projectAPI.DELETE("/data/:collection/:id", dataHandler.DeleteDocument)
		
		// Generic documents endpoints (BaaS standard - alias to /data/{collection})
		projectAPI.GET("/documents/:collection", dataHandler.ListDocuments)
		projectAPI.POST("/documents/:collection", dataHandler.CreateDocument)
		projectAPI.GET("/documents/:collection/:id", dataHandler.GetDocument)
		projectAPI.PUT("/documents/:collection/:id", dataHandler.UpdateDocument)
		projectAPI.DELETE("/documents/:collection/:id", dataHandler.DeleteDocument)
		
		// Advanced document operations (BaaS standard)
		projectAPI.POST("/data/:collection/query", dataHandler.QueryDocuments)
		projectAPI.GET("/data/:collection/count", dataHandler.CountDocuments)
		projectAPI.POST("/data/:collection/batch", dataHandler.BatchCreateDocuments)
		projectAPI.DELETE("/data/:collection/batch", dataHandler.BatchDeleteDocuments)
		
		// Advanced document operations (documents alias)
		projectAPI.POST("/documents/:collection/query", dataHandler.QueryDocuments)
		projectAPI.GET("/documents/:collection/count", dataHandler.CountDocuments)
		projectAPI.POST("/documents/:collection/batch", dataHandler.BatchCreateDocuments)
		projectAPI.DELETE("/documents/:collection/batch", dataHandler.BatchDeleteDocuments)
		
		// Storage management
		projectAPI.GET("/storage/buckets", storageHandler.ListBuckets)
		projectAPI.POST("/storage/buckets", storageHandler.CreateBucket)
		projectAPI.GET("/storage/buckets/:bucket", storageHandler.GetBucket)
		projectAPI.PUT("/storage/buckets/:bucket", storageHandler.UpdateBucket)
		projectAPI.DELETE("/storage/buckets/:bucket", storageHandler.DeleteBucket)
		
		// File management
		projectAPI.GET("/storage/:bucket/files", storageHandler.ListFiles)
		projectAPI.POST("/storage/:bucket/files", storageHandler.UploadFile)
		projectAPI.GET("/storage/:bucket/files/:file_id", storageHandler.GetFile)
		projectAPI.PUT("/storage/:bucket/files/:file_id/move", storageHandler.MoveFile)
		projectAPI.DELETE("/storage/:bucket/files/:file_id", storageHandler.DeleteFile)
		
		// Folder management
		projectAPI.GET("/storage/:bucket/folders", storageHandler.ListFolders)
		projectAPI.POST("/storage/:bucket/folders", storageHandler.CreateFolder)
		projectAPI.DELETE("/storage/:bucket/folders", storageHandler.DeleteFolder)
		
		// Public URL generation for connected apps
		projectAPI.GET("/storage/:bucket/files/:file_id/public-url", storageHandler.GetFilePublicURL)
		projectAPI.POST("/storage/:bucket/files/batch-public-urls", storageHandler.GetBatchFilePublicURLs)
		
		// User management
		projectAPI.GET("/users", userHandler.ListUsers)
		projectAPI.POST("/users", userHandler.CreateUser)
		projectAPI.GET("/users/:user_id", userHandler.GetUser)
		projectAPI.PUT("/users/:user_id", userHandler.UpdateUser)
		projectAPI.DELETE("/users/:user_id", userHandler.DeleteUser)
		
		// User authentication and session management
		projectAPI.POST("/users/logout", userHandler.LogoutUser)
		projectAPI.GET("/users/me", userHandler.GetCurrentUser)
		projectAPI.PUT("/users/:user_id/password", userHandler.ChangePassword)
		projectAPI.GET("/users/:user_id/sessions", userHandler.ListSessions)
		projectAPI.DELETE("/users/:user_id/sessions/:session_id", userHandler.RevokeSession)
		
		// Auth management for project admin interface
		auth := projectAPI.Group("/auth")
		{
			// Auth settings
			auth.GET("/settings", userHandler.GetAuthSettings)
			auth.PUT("/settings", userHandler.UpdateAuthSettings)
			
			// Auth users management (for project admin interface)
			auth.GET("/users", userHandler.ListUsers)
			auth.POST("/users", userHandler.CreateUser)
			auth.PATCH("/users/:user_id", userHandler.UpdateUser)
			auth.DELETE("/users/:user_id", userHandler.DeleteUser)
			
			// Auth providers
			auth.GET("/providers", userHandler.GetAuthProviders)
			auth.PATCH("/providers/:provider_id", userHandler.UpdateAuthProvider)
		}
		
		
		// Functions execution (public access for deployed functions)
		projectAPI.POST("/functions/:function_name", functionHandler.ExecuteFunctionByName)
		projectAPI.GET("/functions/:function_name", functionHandler.ExecuteFunctionByName)
		projectAPI.PUT("/functions/:function_name", functionHandler.ExecuteFunctionByName)
		projectAPI.DELETE("/functions/:function_name", functionHandler.ExecuteFunctionByName)
		
		// Portfolio-specific API endpoints
		portfolio := projectAPI.Group("/")
		{
			// Translations
			portfolio.GET("/translations/languages", portfolioHandler.GetLanguages)
			portfolio.PUT("/translations/languages", portfolioHandler.SetLanguages)
			portfolio.POST("/translations/translate/:pageId", portfolioHandler.TranslatePage)
			portfolio.GET("/translations/page/:pageId", portfolioHandler.GetPageTranslations)
			portfolio.DELETE("/translations/:translationId", portfolioHandler.DeleteTranslation)
			
			// Analytics
			portfolio.GET("/analytics", portfolioHandler.GetAnalytics)
			
			// Images
			portfolio.GET("/images", portfolioHandler.GetImages)
			portfolio.PUT("/images/:id", portfolioHandler.UpdateImage)
			
			// Albums
			portfolio.GET("/albums", portfolioHandler.GetAlbums)
			
			// Pages
			portfolio.GET("/pages", portfolioHandler.GetPages)
			
			// Settings
			portfolio.GET("/settings", portfolioHandler.GetSettings)
			portfolio.PUT("/settings", portfolioHandler.UpdateSettings)
			
			// Branding
			portfolio.GET("/branding", portfolioHandler.GetBranding)
			
			// Portfolio Users
			portfolio.GET("/portfolio/users", portfolioHandler.GetPortfolioUsers)
		}
		
		// Project Templates - Setup and management
		templates := projectAPI.Group("/templates")
		{
			templates.GET("", templateHandler.ListTemplates)
			templates.GET("/:template", templateHandler.GetTemplate)
			templates.POST("/:template/setup", templateHandler.SetupPhotoPortfolio)
		}

		// Template deployment routes
		templateDeployments := projectAPI.Group("/template-deployments")
		{
			templateDeployments.GET("", templateDeploymentHandler.ListTemplateDeployments)
			templateDeployments.POST("", templateDeploymentHandler.CreateTemplateDeployment)
		}

		// Repository compatibility routes
		compatibility := projectAPI.Group("/compatibility")
		{
			compatibility.POST("/check", compatibilityHandler.CheckRepositoryCompatibility)
			compatibility.GET("/repositories/:id", compatibilityHandler.CheckGitHubRepositoryCompatibility)
		}
	}

	// ===========================================
	// PUBLIC PROJECT API ROUTES (/p/{slug}/api/*)
	// No Authentication Required
	// ===========================================
	
	// Public project routes (no authentication required) - registered AFTER protected routes
	projectPublic := r.Group("/p/:project_slug/api")
	projectPublic.Use(middleware.ProjectOnly(cfg, db)) // Only validate project exists
	projectPublic.Use(middleware.ProjectCORS(cfg, db)) // Apply project-specific CORS
	{
		// User authentication (public endpoints)
		projectPublic.POST("/users/login", userHandler.LoginUser)
	}

	// ===========================================
	// STATIC FILE SERVING
	// ===========================================
	
	// Static file serving for deployments
	r.Static("/static", "./uploads/deployments")
	
	// Public file serving for storage buckets
	r.Static("/storage", "./uploads")

	return r
}