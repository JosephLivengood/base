package subscription

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"base/api/internal/domain/organization"
	"base/api/internal/middleware"
	"base/api/pkg/response"
	"base/api/pkg/validate"

	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

type Handler struct {
	repo        *Repository
	orgRepo     *organization.Repository
	stripe      *StripeService
	environment string
}

func NewHandler(repo *Repository, orgRepo *organization.Repository, stripe *StripeService, environment string) *Handler {
	return &Handler{
		repo:        repo,
		orgRepo:     orgRepo,
		stripe:      stripe,
		environment: environment,
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	// Verify user is a member of the org
	_, err := h.orgRepo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, organization.ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}

	sub, err := h.repo.GetActiveByOrganization(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// No active subscription - return null
			response.OK(w, nil)
			return
		}
		response.InternalError(w, "failed to get subscription")
		return
	}

	response.OK(w, sub)
}

// CreateCheckout creates a Stripe Checkout session for subscription upgrade
func (h *Handler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	// Verify user is an admin or owner of the org
	member, err := h.orgRepo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, organization.ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}
	if member.Role != organization.RoleOwner && member.Role != organization.RoleAdmin {
		response.Forbidden(w, "only admins and owners can manage billing")
		return
	}

	// Parse request
	var req CheckoutRequest
	if err := validate.Request(r, &req); err != nil {
		response.BadRequest(w, validate.FirstError(err))
		return
	}

	// Validate price ID
	if req.PriceID != PricePersonalPlusMonthly && req.PriceID != PricePersonalPlusYearly {
		response.BadRequest(w, "invalid price ID")
		return
	}

	// Get organization details
	org, err := h.orgRepo.GetByID(r.Context(), orgID)
	if err != nil {
		response.InternalError(w, "failed to get organization")
		return
	}

	// Get or create Stripe customer
	customer, err := h.stripe.GetOrCreateCustomer(orgID, org.Name, usr.Email)
	if err != nil {
		response.InternalError(w, "failed to create customer")
		return
	}

	// Create checkout session
	session, err := h.stripe.CreateCheckoutSession(customer.ID, req.PriceID, req.SuccessURL, req.CancelURL)
	if err != nil {
		response.InternalError(w, "failed to create checkout session")
		return
	}

	response.OK(w, CheckoutResponse{
		CheckoutURL: session.URL,
	})
}

// CreatePortalSession creates a Stripe Billing Portal session
func (h *Handler) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	// Verify user is an admin or owner of the org
	member, err := h.orgRepo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, organization.ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}
	if member.Role != organization.RoleOwner && member.Role != organization.RoleAdmin {
		response.Forbidden(w, "only admins and owners can manage billing")
		return
	}

	// Parse request
	var req PortalRequest
	if err := validate.Request(r, &req); err != nil {
		response.BadRequest(w, validate.FirstError(err))
		return
	}

	// Get current subscription to find customer ID
	sub, err := h.repo.GetActiveByOrganization(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.BadRequest(w, "no active subscription found")
			return
		}
		response.InternalError(w, "failed to get subscription")
		return
	}

	if sub.StripeCustomerID == nil {
		response.BadRequest(w, "no Stripe customer associated with this subscription")
		return
	}

	// Create billing portal session
	session, err := h.stripe.CreateBillingPortalSession(*sub.StripeCustomerID, req.ReturnURL)
	if err != nil {
		response.InternalError(w, "failed to create portal session")
		return
	}

	response.OK(w, PortalResponse{
		PortalURL: session.URL,
	})
}

// ConfirmCheckout confirms a checkout session and updates the subscription (dev only)
func (h *Handler) ConfirmCheckout(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	// Verify user is an admin or owner of the org
	member, err := h.orgRepo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, organization.ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}
	if member.Role != organization.RoleOwner && member.Role != organization.RoleAdmin {
		response.Forbidden(w, "only admins and owners can manage billing")
		return
	}

	// Parse request
	var req ConfirmRequest
	if err := validate.Request(r, &req); err != nil {
		response.BadRequest(w, validate.FirstError(err))
		return
	}

	// Get checkout session from Stripe
	session, err := h.stripe.GetCheckoutSession(req.SessionID)
	if err != nil {
		response.InternalError(w, "failed to get checkout session")
		return
	}

	// Verify session is complete
	if session.Status != stripe.CheckoutSessionStatusComplete {
		response.BadRequest(w, "checkout session is not complete")
		return
	}

	// Update or create subscription
	sub, err := h.updateSubscriptionFromStripe(r.Context(), orgID, session)
	if err != nil {
		response.InternalError(w, "failed to update subscription")
		return
	}

	response.OK(w, sub)
}

// HandleWebhook handles Stripe webhook events
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		response.BadRequest(w, "error reading request body")
		return
	}

	// In development without webhook secret, just accept the event
	var event stripe.Event
	if h.stripe.WebhookSecret() == "" {
		if err := json.Unmarshal(payload, &event); err != nil {
			response.BadRequest(w, "error parsing webhook JSON")
			return
		}
	} else {
		// Verify webhook signature in production
		sigHeader := r.Header.Get("Stripe-Signature")
		event, err = webhook.ConstructEvent(payload, sigHeader, h.stripe.WebhookSecret())
		if err != nil {
			response.BadRequest(w, "webhook signature verification failed")
			return
		}
	}

	// Handle the event
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			response.BadRequest(w, "error parsing session")
			return
		}
		// Get full session with subscription details
		fullSession, err := h.stripe.GetCheckoutSession(session.ID)
		if err != nil {
			response.InternalError(w, "error getting session")
			return
		}
		// Get org ID from customer metadata
		if fullSession.Customer != nil {
			orgID := fullSession.Customer.Metadata["organization_id"]
			if orgID != "" {
				_, err = h.updateSubscriptionFromStripe(r.Context(), orgID, fullSession)
				if err != nil {
					response.InternalError(w, "error updating subscription")
					return
				}
			}
		}

	case "customer.subscription.updated", "customer.subscription.deleted":
		var stripeSub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
			response.BadRequest(w, "error parsing subscription")
			return
		}
		// Find subscription by Stripe subscription ID
		sub, err := h.repo.GetByStripeSubscriptionID(r.Context(), stripeSub.ID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				// Subscription not found in our system, ignore
				w.WriteHeader(http.StatusOK)
				return
			}
			response.InternalError(w, "error finding subscription")
			return
		}
		// Update subscription status
		status := mapStripeStatus(stripeSub.Status)
		var activeUntil *time.Time
		if stripeSub.CurrentPeriodEnd > 0 {
			t := time.Unix(stripeSub.CurrentPeriodEnd, 0)
			activeUntil = &t
		}
		_, err = h.repo.UpdateStripeInfo(r.Context(), sub.ID,
			sub.StripeCustomerID, sub.StripeSubscriptionID, sub.StripePriceID,
			sub.Tier, status, activeUntil)
		if err != nil {
			response.InternalError(w, "error updating subscription")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) updateSubscriptionFromStripe(ctx context.Context, orgID string, session *stripe.CheckoutSession) (*Subscription, error) {
	customerID := ""
	if session.Customer != nil {
		customerID = session.Customer.ID
	}

	subID := ""
	priceID := ""
	var activeUntil *time.Time

	if session.Subscription != nil {
		subID = session.Subscription.ID
		if len(session.Subscription.Items.Data) > 0 {
			priceID = session.Subscription.Items.Data[0].Price.ID
		}
		if session.Subscription.CurrentPeriodEnd > 0 {
			t := time.Unix(session.Subscription.CurrentPeriodEnd, 0)
			activeUntil = &t
		}
	}

	// Check if subscription already exists for this org
	existing, err := h.repo.GetActiveByOrganization(ctx, orgID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	if existing != nil {
		// Update existing subscription
		return h.repo.UpdateStripeInfo(ctx, existing.ID,
			&customerID, &subID, &priceID,
			TierPersonalPlus, StatusActive, activeUntil)
	}

	// Create new subscription
	return h.repo.CreateFromStripe(ctx, orgID, TierPersonalPlus, StatusActive,
		&customerID, &subID, &priceID,
		time.Now(), activeUntil)
}

func mapStripeStatus(status stripe.SubscriptionStatus) string {
	switch status {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing:
		return StatusActive
	case stripe.SubscriptionStatusCanceled, stripe.SubscriptionStatusUnpaid:
		return StatusCanceled
	case stripe.SubscriptionStatusPastDue:
		return StatusPastDue
	default:
		return StatusActive
	}
}
