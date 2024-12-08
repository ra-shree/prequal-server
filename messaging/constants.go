package messaging

const PUBLISHING_QUEUE string = "reverseproxy-to-admin"
const CONSUMING_QUEUE string = "admin-to-reverseproxy"

// for consuming
const ADD_REPLICA string = "add-replica"
const REMOVE_REPLICA string = "remove-replica"

// for publishing
const ADDED_REPLICA string = "replica-added"
const REMOVED_REPLICA string = "replica-removed"
const REPLICA_ADD_FAILED string = "replica-add-failed"
const STATISTICS string = "statistics"
const PARAMETERS_UPDATED = "parameters-updated"
const PARAMETERS_UPDATE_FAILED = "parameters-update-failed"
