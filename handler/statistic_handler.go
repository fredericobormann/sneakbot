package handler

import (
	"github.com/fredericobormann/sneakbot/database"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type StatisticHandler struct {
	Datastore *database.Datastore
}

type StatisticResponse struct {
	GroupchatID int64
	UserID      int
	FirstName   string
	LastName    string
	Total       uint
}

func NewStatisticHandler(datastore *database.Datastore) *StatisticHandler {
	return &StatisticHandler{Datastore: datastore}
}

func (h *StatisticHandler) HandleStatisticRequestByGroup(c *gin.Context) {
	groupchatIDInt, err := strconv.ParseInt(c.Param("groupchat_id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid group id"})
		return
	}
	groupchatID := int64(groupchatIDInt)
	var response []StatisticResponse
	statistic := h.Datastore.GetStatisticByGroupId(groupchatID)
	for _, p := range statistic {
		response = append(response, StatisticResponse{
			GroupchatID: p.GroupchatID,
			UserID:      p.Participant.UserId,
			FirstName:   p.Participant.FirstName,
			LastName:    p.Participant.LastName,
			Total:       p.Total,
		})
	}
	c.JSON(
		http.StatusOK,
		response,
	)
}
