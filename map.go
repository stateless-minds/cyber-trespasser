package main

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

const dbRoute = "route"

type Status string

const (
	StatusUnverified Status = "unverified"
	StatusVerified   Status = "verified"
)

// delivery is the delivery component that displays a delivery on a map. A component is a
// customizable, independent, and reusable UI elemend. It is created by
// embedding app.Compo into a strucd.
type mapLibre struct {
	app.Compo
	notificationPermission app.NotificationPermission
	sh                     *shell.Shell
	routesJSON             string
	myPeerID               string
}

type Route struct {
	ID          string     `mapstructure:"_id" json:"_id" validate:"uuid_rfc4122"`                 // Unique identifier for the delivery
	Coordinates [][]string `mapstructure:"coordinates" json:"coordinates" validate:"uuid_rfc4122"` // Coordinates
	CreatedBy   string     `mapstructure:"created_by" json:"created_by" validate:"uuid_rfc4122"`   // Created By
	Verifiers   []string   `mapstructure:"verifiers" json:"verifiers" validate:"uuid_rfc4122"`     // Verifiers
	Status      Status     `mapstructure:"status" json:"status" validate:"uuid_rfc4122"`           // Status
}

func (m *mapLibre) OnMount(ctx app.Context) {
	m.notificationPermission = ctx.Notifications().Permission()
	switch m.notificationPermission {
	case app.NotificationDefault:
		m.notificationPermission = ctx.Notifications().RequestPermission()
	case app.NotificationDenied:
		app.Window().Call("alert", "In order to use Cyber Trespasser notifications needs to be enabled")
		return
	}

	sh := shell.NewShell("localhost:5001")
	m.sh = sh

	myPeer, err := m.sh.ID()
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
	}

	m.myPeerID = myPeer.ID

	// get all routes
	m.getRoutes(ctx)

	app.Window().Call("setupMap", m.myPeerID, m.routesJSON)

	m.setupRouteCreatedListener(ctx)

	m.setupRouteVerifiedListener(ctx)
}

func (m *mapLibre) getRoutes(ctx app.Context) {
	routesJSON, err := m.sh.OrbitDocsQuery(dbRoute, "all", "")
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	if strings.TrimSpace(string(routesJSON)) != "null" && len(routesJSON) > 0 {
		m.routesJSON = string(routesJSON)
	}
}

func (m *mapLibre) setupRouteCreatedListener(ctx app.Context) {
	app.Window().GetElementByID("map").Call("addEventListener", "route-created", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		event := args[0]

		dc := event.Get("detail").Get("latlngs").String()

		app.Log(dc)

		// Now unmarshal the JSON from the cleaned string
		var deliveryCoordinates [][]string
		err := json.Unmarshal([]byte(dc), &deliveryCoordinates)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		delivery := Route{
			ID:          uuid.NewString(),
			Coordinates: deliveryCoordinates,
			CreatedBy:   m.myPeerID,
			Verifiers:   []string{},
			Status:      StatusUnverified,
		}

		deliveryJSON, err := json.Marshal(delivery)

		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
		}

		ctx.Async(func() {
			err = m.sh.OrbitDocsPut(dbRoute, deliveryJSON)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
			}

			ctx.Dispatch(func(ctx app.Context) {
				ctx.Notifications().New(app.Notification{
					Title: "Success",
					Body:  "Route was saved. Once it gets verified by 3 peers you will get access to the Cyber Trespasser network.",
				})
				ctx.Reload()
			})
		})

		return nil
	}))
}

func (m *mapLibre) setupRouteVerifiedListener(ctx app.Context) {
	app.Window().GetElementByID("map").Call("addEventListener", "route-verified", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		event := args[0]

		routeID, err := strconv.Unquote(event.Get("detail").Get("id").String())
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		var routes []Route

		err = json.Unmarshal([]byte(m.routesJSON), &routes)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return nil
		}

		for _, route := range routes {
			if route.ID == routeID {
				route.Verifiers = append(route.Verifiers, m.myPeerID)
				if len(route.Verifiers) == 3 {
					route.Status = StatusVerified
				}

				routeJSON, err := json.Marshal(route)

				if err != nil {
					ctx.Notifications().New(app.Notification{
						Title: "Error",
						Body:  err.Error(),
					})
				}

				ctx.Async(func() {
					err = m.sh.OrbitDocsPut(dbRoute, routeJSON)
					if err != nil {
						ctx.Notifications().New(app.Notification{
							Title: "Error",
							Body:  err.Error(),
						})
					}

					ctx.Dispatch(func(ctx app.Context) {
						ctx.Notifications().New(app.Notification{
							Title: "Success",
							Body:  "Route was saved",
						})
						ctx.Reload()
					})
				})
			}
		}

		return nil
	}))
}

// The Render method is where the component appearance is defined.
func (m *mapLibre) Render() app.UI {
	return app.Div().ID("map")
}
