package api

import (
	"errors"
	"net/http"
	"time"

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

	var body ruleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	rule := ruleFromBody(body)
	rule.BucketID = bucketID

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

	var body ruleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	rule := ruleFromBody(body)
	rule.RuleID = ruleID

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

type ruleBody struct {
	Name                string   `json:"name"`
	Priority            int32    `json:"priority"`
	DescriptionContains *string  `json:"description_contains"`
	MinAmountAUD        *float64 `json:"min_amount_aud"`
	MaxAmountAUD        *float64 `json:"max_amount_aud"`
	TransactionType     *string  `json:"transaction_type"`
	CategoryID          *string  `json:"category_id"`
	DateFrom            *string  `json:"date_from"`
	DateTo              *string  `json:"date_to"`
	ForeignCurrencyCode *string  `json:"foreign_currency_code"`
}

func ruleFromBody(body ruleBody) service.Rule {
	rule := service.Rule{
		Name:                body.Name,
		Priority:            body.Priority,
		DescriptionContains: body.DescriptionContains,
		TransactionType:     body.TransactionType,
		CategoryID:          body.CategoryID,
		ForeignCurrencyCode: body.ForeignCurrencyCode,
	}
	if body.MinAmountAUD != nil {
		v := int64(*body.MinAmountAUD * 100)
		rule.MinAmountCents = &v
	}
	if body.MaxAmountAUD != nil {
		v := int64(*body.MaxAmountAUD * 100)
		rule.MaxAmountCents = &v
	}
	if body.DateFrom != nil {
		if t, err := time.Parse("2006-01-02", *body.DateFrom); err == nil {
			rule.DateFrom = &t
		}
	}
	if body.DateTo != nil {
		if t, err := time.Parse("2006-01-02", *body.DateTo); err == nil {
			rule.DateTo = &t
		}
	}
	return rule
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
