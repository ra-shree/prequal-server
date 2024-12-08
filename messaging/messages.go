package messaging

import "github.com/ra-shree/prequal-server/common"

func ReplicaRemoved(url string) {
	// sending message to admin server when removing replica
	msg := Message{
		Name: REMOVED_REPLICA,
		Body: url,
	}

	PublishMessage(PUBLISHING_QUEUE, &msg)
}

func ReplicaAdded(url string) {
	// sending message to admin server when adding replica
	msg := Message{
		Name: ADDED_REPLICA,
		Body: url,
	}

	PublishMessage(PUBLISHING_QUEUE, &msg)
}

func ReplicaAddFailed(url string) {
	// sending message to admin server when adding replica fails
	msg := Message{
		Name: REPLICA_ADD_FAILED,
		Body: url,
	}

	PublishMessage(PUBLISHING_QUEUE, &msg)
}

func StatisticsUpdated() {
	jsonData := common.TransformMapToJson()
	msg := Message{
		Name: STATISTICS,
		Body: jsonData,
	}

	PublishMessage(PUBLISHING_QUEUE, &msg)
}

func ParametersUpdated() {
	msg := Message{
		Name: PARAMETERS_UPDATED,
		Body: common.CurrentPrequalParameters,
	}

	PublishMessage(PUBLISHING_QUEUE, &msg)
}

func ParametersUpdateFailed() {
	msg := Message{
		Name: PARAMETERS_UPDATE_FAILED,
		Body: common.CurrentPrequalParameters,
	}

	PublishMessage(PUBLISHING_QUEUE, &msg)
}
