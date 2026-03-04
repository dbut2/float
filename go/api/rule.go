package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dbut.dev/float/go/middleware"
	"dbut.dev/float/go/service"
)

func (a *API) listRules(c *gin.Context) {
	userID := middleware.GetUserID(c)

	rules, err := a.rules.ListRules(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}

func (a *API) listBucketRules(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket ID"})
		return
	}

	rules, err := a.rules.ListRulesByBucket(c.Request.Context(), bucketID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}

func (a *API) createRule(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket ID"})
		return
	}

	var body struct {
		Name                string   `json:"name"`
		Priority            int32    `json:"priority"`
		DescriptionContains *string  `json:"description_contains"`
		MinAmountAUD        *float64 `json:"min_amount_aud"`
		MaxAmountAUD        *float64 `json:"max_amount_aud"`
		TransactionType     *string  `json:"transaction_type"`
		CategoryID          *string  `json:"category_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	rule := service.Rule{
		BucketID:            bucketID,
		Name:                body.Name,
		Priority:            body.Priority,
		DescriptionContains: body.DescriptionContains,
		TransactionType:     body.TransactionType,
		CategoryID:          body.CategoryID,
	}
	if body.MinAmountAUD != nil {
		v := int64(*body.MinAmountAUD * 100)
		rule.MinAmountCents = &v
	}
	if body.MaxAmountAUD != nil {
		v := int64(*body.MaxAmountAUD * 100)
		rule.MaxAmountCents = &v
	}

	created, err := a.rules.CreateRule(c.Request.Context(), rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = userID
	c.JSON(http.StatusCreated, created)
}

func (a *API) updateRule(c *gin.Context) {
	userID := middleware.GetUserID(c)

	ruleID, err := uuid.Parse(c.Param("ruleID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule ID"})
		return
	}

	var body struct {
		Name                string   `json:"name"`
		Priority            int32    `json:"priority"`
		DescriptionContains *string  `json:"description_contains"`
		MinAmountAUD        *float64 `json:"min_amount_aud"`
		MaxAmountAUD        *float64 `json:"max_amount_aud"`
		TransactionType     *string  `json:"transaction_type"`
		CategoryID          *string  `json:"category_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	rule := service.Rule{
		RuleID:              ruleID,
		Name:                body.Name,
		Priority:            body.Priority,
		DescriptionContains: body.DescriptionContains,
		TransactionType:     body.TransactionType,
		CategoryID:          body.CategoryID,
	}
	if body.MinAmountAUD != nil {
		v := int64(*body.MinAmountAUD * 100)
		rule.MinAmountCents = &v
	}
	if body.MaxAmountAUD != nil {
		v := int64(*body.MaxAmountAUD * 100)
		rule.MaxAmountCents = &v
	}

	updated, err := a.rules.UpdateRule(c.Request.Context(), rule, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (a *API) deleteRule(c *gin.Context) {
	userID := middleware.GetUserID(c)

	ruleID, err := uuid.Parse(c.Param("ruleID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule ID"})
		return
	}

	if err := a.rules.DeleteRule(c.Request.Context(), ruleID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *API) applyRulesToGeneral(c *gin.Context) {
	userID := middleware.GetUserID(c)

	applied, err := a.rules.ApplyRulesToGeneral(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"applied": applied})
}
