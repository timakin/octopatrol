package client

import (
	"encoding/json"
	"log"
	"sort"

	"github.com/google/go-github/github"
	"github.com/patrickmn/go-cache"
)

func (i Instance) GetNotifications() []github.Notification {
	if cv, found := i.cache.Get("notifications"); found {
		cachedNotifications := cv.([]github.Notification)
		return cachedNotifications
	}
	opt := &github.NotificationListOptions{All: true}
	notifications, _, err := i.ghCli.Activity.ListNotifications(opt)
	if err != nil {
		log.Fatal(err)
	}
	i.cache.Set("notifications", notifications, cache.DefaultExpiration)
	return notifications
}

func (i Instance) GetIssues(owner string, repo string) []github.Issue {
	opt := &github.IssueListByRepoOptions{State: "open"}
	issues, _, err := i.ghCli.Issues.ListByRepo(owner, repo, opt)
	if err != nil {
		log.Fatal(err)
	}
	return issues
}

func (i Instance) GetPullRequests(owner string, repo string) []*github.PullRequestEvent {
	opt := &github.ListOptions{PerPage: 100}
	events, _, err := i.ghCli.Activity.ListRepositoryEvents(owner, repo, opt)
	if err != nil {
		log.Fatal(err)
	}

	var pullRequestEvents []*github.PullRequestEvent
	for i, event := range events {
		var pullreqPayload *github.PullRequestEvent
		err := json.Unmarshal(*event.RawPayload, &pullreqPayload)
		if err != nil {
			panic(err)
		}
		pullRequestEvents[i] = pullreqPayload
	}

	pullreqs := PullReqFilter(events, func(e github.PullRequestEvent) bool {
		isOpen := *e.PullRequest.State == "open"
		isValidType := *e.Type == "PullRequestEvent"
		return isOpen && isValidType
	})
	return pullreqs
}

func (i Instance) GetRepoNotificationCounters() RepoNotificationCounters {
	repos := i.GetListFollowingRepository()
	repoNotificationCounters := make(RepoNotificationCounters, len(repos))
	for index, repo := range repos {
		repo := repo
		unreadCount := i.countUnreadRepositoryNotification(repo.Owner.Login, repo.Name)
		repoNotificationCounter := &RepoNotificationCounter{
			Repository:              &repo,
			UnreadNotificationCount: unreadCount,
		}
		repoNotificationCounters[index] = repoNotificationCounter
	}
	sort.Sort(repoNotificationCounters)
	return repoNotificationCounters
}

func (i Instance) GetListFollowingRepository() []github.Repository {
	opt := &github.ListOptions{PerPage: 100}
	userId := i.getAuthenticatedUserId()
	Repositories, _, err := i.ghCli.Activity.ListWatched(*userId, opt)
	if err != nil {
		log.Fatal(err)
	}
	return Repositories
}

func (i Instance) getAuthenticatedUserId() *string {
	User, _, err := i.ghCli.Users.Get("")
	if err != nil {
		log.Fatal(err)
	}
	return User.Login
}

func (i Instance) countUnreadRepositoryNotification(owner *string, repoName *string) int {
	notifications := i.GetNotifications()
	unreadRepositoryNotifications := NotificationFilter(notifications, func(n github.Notification) bool {
		return *n.Repository.Owner.Login == *owner && *n.Repository.Name == *repoName
	})
	return len(unreadRepositoryNotifications)
}
