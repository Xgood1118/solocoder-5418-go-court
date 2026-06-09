package main

import (
	"fmt"
	"os"
	"time"

	"court-scheduler/handlers"
	"court-scheduler/services"
	"court-scheduler/store"

	"github.com/gin-gonic/gin"
)

func main() {
	snapshotDir := getEnv("SNAPSHOT_DIR", "./snapshots")
	port := getEnv("PORT", "8080")

	store.InitStore(snapshotDir)
	store.GlobalStore.StartAutoSnapshot(1 * time.Hour)

	services.StartFollowUpScheduler()
	services.StartMonthlyStatsScheduler()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	courtroomRoutes := r.Group("/courtrooms")
	{
		courtroomRoutes.GET("", handlers.ListCourtrooms)
		courtroomRoutes.POST("", handlers.CreateCourtroom)
		courtroomRoutes.GET("/:id", handlers.GetCourtroom)
		courtroomRoutes.PUT("/:id", handlers.UpdateCourtroom)
		courtroomRoutes.DELETE("/:id", handlers.DeleteCourtroom)
	}

	judgeRoutes := r.Group("/judges")
	{
		judgeRoutes.GET("", handlers.ListJudges)
		judgeRoutes.POST("", handlers.CreateJudge)
		judgeRoutes.GET("/:id", handlers.GetJudge)
		judgeRoutes.PUT("/:id", handlers.UpdateJudge)
		judgeRoutes.DELETE("/:id", handlers.DeleteJudge)
		judgeRoutes.GET("/:id/daily-count", handlers.GetJudgeDailyCount)
	}

	jurorRoutes := r.Group("/jurors")
	{
		jurorRoutes.GET("", handlers.ListJurors)
		jurorRoutes.POST("", handlers.CreateJuror)
		jurorRoutes.GET("/:id", handlers.GetJuror)
		jurorRoutes.PUT("/:id", handlers.UpdateJuror)
		jurorRoutes.DELETE("/:id", handlers.DeleteJuror)
	}

	caseRoutes := r.Group("/cases")
	{
		caseRoutes.GET("", handlers.ListCases)
		caseRoutes.POST("", handlers.CreateCase)
		caseRoutes.GET("/:id", handlers.GetCase)
		caseRoutes.PUT("/:id", handlers.UpdateCase)
		caseRoutes.DELETE("/:id", handlers.DeleteCase)
	}

	hearingRoutes := r.Group("/hearings")
	{
		hearingRoutes.GET("", handlers.ListHearings)
		hearingRoutes.POST("", handlers.CreateHearing)
		hearingRoutes.GET("/:id", handlers.GetHearing)
		hearingRoutes.POST("/:id/postpone", handlers.PostponeHearing)
		hearingRoutes.POST("/:id/cancel", handlers.CancelHearing)
		hearingRoutes.POST("/:id/complete", handlers.CompleteHearing)
	}

	noticeRoutes := r.Group("/notices")
	{
		noticeRoutes.GET("", handlers.ListNotices)
		noticeRoutes.GET("/types", handlers.GetNoticeTypes)
		noticeRoutes.GET("/:id", handlers.GetNotice)
		noticeRoutes.PUT("/:id/status", handlers.UpdateNoticeStatus)
	}

	statsRoutes := r.Group("/stats")
	{
		statsRoutes.GET("", handlers.ListStats)
		statsRoutes.GET("/:month", handlers.GetMonthlyStats)
		statsRoutes.POST("/generate", handlers.GenerateStats)
	}

	fmt.Printf("Court Scheduler API starting on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
