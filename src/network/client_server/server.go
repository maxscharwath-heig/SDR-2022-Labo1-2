package client_server

import (
	"io"
	"net"
	"sdr/labo1/src/network"
	"sdr/labo1/src/types"
	"sdr/labo1/src/utils"
)

// ServerEndpoint extends the Endpoint struct with a function that is called when the endpoint is called.
type ServerEndpoint = Endpoint[HeaderResponse]

// AuthId
// The type of the auth object
type AuthId = int

// HeaderResponse
// is the first response of the server.
// - Valid: true if the endpoint is valid
// - NeedsAuth: true if the endpoint needs authentication
type HeaderResponse struct {
	Valid     bool   `json:"valid"`
	NeedsAuth bool   `json:"needsAuth"`
	AuthId    AuthId `json:"-"`
}

type pendingRequest struct {
	request  network.Request[HeaderResponse]
	callback func()
}

// ServerProtocol
// is the protocol that is used to handle the server side of the protocol.
// - AuthFunc: the function that is called to authenticate the user.
// - Endpoints: the endpoints that are registered. It is a map of the endpointId and the endpoint.
type ServerProtocol struct {
	AuthFunc       AuthFunc
	Endpoints      map[string]ServerEndpoint
	pendingRequest chan pendingRequest
}

func CreateServerProtocol(authFunc AuthFunc, endpoints map[string]ServerEndpoint) ServerProtocol {
	return ServerProtocol{
		AuthFunc:       authFunc,
		Endpoints:      endpoints,
		pendingRequest: make(chan pendingRequest, 1),
	}
}

func (p ServerProtocol) ProcessRequests() {
	for {
		select {
		case pending := <-p.pendingRequest:
			utils.LogInfo(true, "Start processing request", pending.request.EndpointId, pending.request.Conn.RemoteAddr())
			pending.callback()
			utils.LogInfo(true, "Finished processing request", pending.request.EndpointId, pending.request.Conn.RemoteAddr())
		}
	}
}

// HandleConnection is the function that is called to process the connection. It is called in a go routine.
func (p ServerProtocol) HandleConnection(c net.Conn) {
	utils.LogInfo(false, "new connection", c.RemoteAddr())
	defer func() {
		utils.LogInfo(true, "close connection", c.RemoteAddr())
		_ = c.Close()
	}()

	conn := network.CreateConnection(c)
	var err error
	ready := make(chan struct{}, 1)
	ready <- struct{}{} // A connection can handle one request at a time
	for {
		if conn.IsClosed() || err == io.EOF {
			utils.LogInfo(false, "connection closed", c.RemoteAddr())
			break
		}

		select {
		case <-ready:

			request := network.Request[HeaderResponse]{Conn: c}

			request.EndpointId, err = conn.GetLine()
			if err != nil {
				utils.LogInfo(false, "error while receiving endpointId", err)
				continue
			}

			endpoint, ok := p.Endpoints[request.EndpointId]
			if ok {
				request.Header.Valid = true
				request.Header.NeedsAuth = endpoint.NeedsAuth
			}

			err = conn.SendJSON(request.Header)
			if err != nil {
				utils.LogWarning(false, "error while sending header", err)
				continue
			}

			if !request.Header.Valid {
				utils.LogWarning(false, "invalid endpoint, canceling request")
				continue
			}

			p.pendingRequest <- pendingRequest{request, func() { // Add the request to the pending requests
				if request.Header.NeedsAuth {
					var credentials types.Credentials

					err = conn.GetJson(&credentials)
					if err != nil {
						utils.LogWarning(false, "error while receiving credentials", err)
						ready <- struct{}{}
						return
					}

					isValid, auth := p.AuthFunc(credentials)

					err = conn.SendJSON(AuthResponse{Success: isValid, Auth: auth})
					if err != nil {
						utils.LogWarning(false, "error while sending auth response", err)
						ready <- struct{}{}
						return
					}

					request.Header.AuthId = auth
					if !isValid {
						utils.LogWarning(false, "invalid credentials, canceling request")
						ready <- struct{}{}
						return
					}
				}
				p.pendingRequest <- pendingRequest{request, func() { // Add processing to the pending request queue
					defer func() {
						ready <- struct{}{}
					}()
					request.Data, err = conn.GetLine()
					if err != nil {
						utils.LogWarning(false, "error while receiving data", err)
						return
					}

					response := endpoint.HandlerFunc(request)
					err = conn.SendJSON(response)
					if err != nil {
						utils.LogWarning(false, "error while sending response", err)
						return
					}
				},
				}
			},
			}
		}
	}
}
