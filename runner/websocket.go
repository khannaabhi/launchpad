package runner

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/spaceuptech/launchpad/model"
	"github.com/spaceuptech/launchpad/utils"
)

var upgrader = websocket.Upgrader{}

func (runner *Runner) handleWebsocketRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.Errorln("Could not upgrade to websocket (autoscaler):", err)
			return
		}
		defer utils.CloseReaderCloser(c)

		// Check if token is valid
		claims, err := runner.auth.VerifyProxyToken(utils.GetToken(r))
		if err != nil {
			logrus.Errorf("Failed to verify autoscaler socket connection - %s", err.Error())
			return
		}

		// Extract node id, project id and service name
		nodeIDTemp, ok1 := claims["id"]
		projectTemp, ok2 := claims["project"]
		serviceTemp, ok3 := claims["service"]
		versionTemp, ok4 := claims["version"]
		if !ok1 || !ok2 || !ok3 || !ok4 {
			logrus.Errorln("Failed to establish autoscaler socket connection - token does not contain valid claims")
			return
		}

		nodeID := nodeIDTemp.(string)
		project := projectTemp.(string)
		service := serviceTemp.(string)
		version := versionTemp.(string)

		for {
			msg := new(model.ProxyMessage)
			if err := c.ReadJSON(msg); err != nil {
				logrus.Errorf("Failed to receive message from proxy (%s:%s): %s", project, service, err.Error())
				return
			}

			// Set crucial meta data
			msg.NodeID = nodeID
			msg.Project = project
			msg.Service = service
			msg.Version = version

			// Append msg to disk
			runner.chAppend <- msg
		}
	}
}
