const Mutation = `
type Mutation {
  ping: String
  createGithubNonce: String!
  createGithubAuthToken(state: String!, code: String!): AccessToken
  trackScmLead(deploymentPreference: String!, emailAddress: String!, scmProvider: String!): String
  logout: String

  createShipOpsCluster(title: String!): Cluster
  createGitOpsCluster(title: String!, installationId: Int, gitOpsRef: GitOpsRefInput): Cluster
  updateCluster(clusterId: String!, clusterName: String!, gitOpsRef: GitOpsRefInput): Cluster
  deleteCluster(clusterId: String!): Boolean

  createWatch(stateJSON: String!, owner: String!, clusterID: String, githubPath: String): Watch
  updateWatch(watchId: String!, watchName: String, iconUri: String): Watch
  deleteWatch(watchId: String!, childWatchIds: [String]): Boolean
  updateStateJSON(slug: String!, stateJSON: String!): Watch
  deployWatchVersion(watchId: String!, sequence: Int): Boolean
  saveWatchContributors(id: String!, contributors: [ContributorInput]!): [Contributor]

  createNotification(watchId: String!, webhook: WebhookNotificationInput, email: EmailNotificationInput): Notification
  updateNotification(watchId: String!, notificationId: String!, webhook: WebhookNotificationInput, email: EmailNotificationInput): Notification
  enableNotification(watchId: String!, notificationId: String!, enabled: Int!): Notification
  deleteNotification(id: String!, isPending: Boolean): Boolean

  createInitSession(pendingInitId: String, upstreamUri: String, clusterID: String, githubPath: String): InitSession!
  createUnforkSession(upstreamUri: String!, forkUri: String!): UnforkSession!
  createUpdateSession(watchId: ID!): UpdateSession!
  createEditSession(watchId: ID!): EditSession!

  uploadImageWatchBatch(imageList: String!): String
}
`;

export const all = [Mutation];
