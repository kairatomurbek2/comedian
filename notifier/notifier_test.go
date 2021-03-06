package notifier

import (
	"fmt"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

type ChatStub struct {
	LastMessage string
}

func (c *ChatStub) Run() error {
	return nil
}

func (c *ChatStub) SendMessage(chatID, message string) error {
	c.LastMessage = fmt.Sprintf("CHAT: %s, MESSAGE: %s", chatID, message)
	return nil
}

func (c *ChatStub) SendUserMessage(userID, message string) error {
	c.LastMessage = fmt.Sprintf("CHAT: %s, MESSAGE: %s", userID, message)
	return nil
}

func TestNotifier(t *testing.T) {
	c, err := config.Get()
	c.ReminderRepeatsMax = 0
	c.ReminderTime = 0
	c.NotifierInterval = 0
	assert.NoError(t, err)
	ch := &ChatStub{}
	n, err := NewNotifier(c, ch)
	assert.NoError(t, err)

	channelID := "QWERTY123"

	d := time.Date(2018, 1, 2, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	su, err := n.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   channelID,
		Channel:     "chanName",
		Role:        "user",
	})
	assert.NoError(t, err)
	fmt.Println(su.Created)
	su2, err := n.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID2",
		SlackName:   "user2",
		ChannelID:   channelID,
		Channel:     "chanName",
		Role:        "user",
	})
	assert.NoError(t, err)
	fmt.Println(su2.Created)
	nonReporters, err := n.getCurrentDayNonReporters(channelID)
	assert.NoError(t, err)
	assert.NotEmpty(t, nonReporters)
	assert.Equal(t, 2, len(nonReporters))

	n.SendWarning(channelID)
	assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Hey, <@userID1>, <@userID2>! 0 minutes to deadline and the team is still waiting for standups from you!", ch.LastMessage)

	n.SendChannelNotification(channelID)
	assert.Equal(t, "CHAT: userID2, MESSAGE: Hello, <@user2>! You missed the standup deadline in <#QWERTY123> channel. Please, write you standup ASAP!", ch.LastMessage)

	n.NotifyChannels()
	assert.Equal(t, "CHAT: userID2, MESSAGE: Hello, <@user2>! You missed the standup deadline in <#QWERTY123> channel. Please, write you standup ASAP!", ch.LastMessage)

	d = time.Date(2018, 1, 2, 9, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	s, err := n.DB.CreateStandup(model.Standup{
		Created:    time.Now(),
		Modified:   time.Now(),
		ChannelID:  channelID,
		Comment:    "work hard",
		UsernameID: "userID1",
		Username:   "user1",
		MessageTS:  "qweasdzxc",
	})
	assert.NoError(t, err)

	// add standup for user @user2
	s2, err := n.DB.CreateStandup(model.Standup{
		Created:    time.Now(),
		Modified:   time.Now(),
		ChannelID:  channelID,
		Comment:    "hello world",
		UsernameID: "userID2",
		Username:   "user2",
		MessageTS:  "qweasd",
	})

	d = time.Date(2018, 1, 2, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	nonReporters, err = n.getCurrentDayNonReporters(channelID)
	assert.NoError(t, err)
	assert.Empty(t, nonReporters)

	n.SendChannelNotification(channelID)
	assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Congradulations! Everybody wrote their standups today!", ch.LastMessage)

	assert.NoError(t, n.DB.DeleteStandupUser(su.SlackName, su.ChannelID))
	assert.NoError(t, n.DB.DeleteStandupUser(su2.SlackName, su2.ChannelID))

	assert.NoError(t, n.DB.DeleteStandup(s.ID))
	assert.NoError(t, n.DB.DeleteStandup(s2.ID))
}

func TestCheckUser(t *testing.T) {
	c, err := config.Get()
	c.ChanGeneral = "XXXYYYZZZ"
	assert.NoError(t, err)
	ch := &ChatStub{}
	n, err := NewNotifier(c, ch)
	assert.NoError(t, err)

	users, err := n.DB.ListAllStandupUsers()
	assert.NoError(t, err)
	for _, user := range users {
		assert.NoError(t, n.DB.DeleteStandupUser(user.SlackName, user.ChannelID))
	}

	d := time.Date(2018, 6, 24, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	channelID := "QWERTY123"
	st, err := n.DB.CreateStandupTime(model.StandupTime{
		ChannelID: channelID,
		Channel:   "chanName",
		Time:      time.Now().Unix(),
	})

	d = time.Date(2018, 6, 25, 0, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	u1, err := n.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   channelID,
		Channel:     "chanName",
	})
	assert.NoError(t, err)
	u2, err := n.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID2",
		SlackName:   "user2",
		ChannelID:   channelID,
		Channel:     "chanName",
	})
	assert.NoError(t, err)

	u3, err := n.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID3",
		SlackName:   "user3",
		ChannelID:   channelID,
		Channel:     "chanName",
	})
	assert.NoError(t, err)
	u4, err := n.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID4",
		SlackName:   "user4",
		ChannelID:   channelID,
		Channel:     "chanName",
	})
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/rest/api/v1/logger/users/userID1/2018-06-25/2018-06-25", c.CollectorURL),
		httpmock.NewStringResponder(200, `{"total_commits": 2, "total_merges": 1, "worklogs": 100000}`))

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/rest/api/v1/logger/users/userID2/2018-06-25/2018-06-25", c.CollectorURL),
		httpmock.NewStringResponder(200, `{"total_commits": 30, "total_merges": 0, "worklogs": 13600}`))

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/rest/api/v1/logger/users/userID3/2018-06-25/2018-06-25", c.CollectorURL),
		httpmock.NewStringResponder(200, `{"total_commits": 0, "total_merges": 0, "worklogs": 0}`))

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/rest/api/v1/logger/users/userID4/2018-06-25/2018-06-25", c.CollectorURL),
		httpmock.NewStringResponder(200, `{"total_commits": 20, "total_merges": 0, "worklogs": 50000}`))

	testCases := []struct {
		title         string
		user          model.StandupUser
		worklogs      int
		commits       int
		isNonReporter bool
		err           error
	}{
		{"test 1", u1, 27, 2, true, nil},
		{"test 2", u2, 3, 30, true, nil},
		{"test 3", u3, 0, 0, true, nil},
		{"test 4", u4, 13, 20, true, nil},
	}

	for _, tt := range testCases {
		worklogs, commits, err := n.getCollectorData(tt.user, time.Now(), time.Now())
		assert.NoError(t, err)
		assert.Equal(t, tt.worklogs, worklogs)
		assert.Equal(t, tt.commits, commits)
		isNonReporter, err := n.DB.IsNonReporter(tt.user.SlackUserID, tt.user.ChannelID, time.Now(), time.Now())
		assert.NoError(t, err)
		assert.Equal(t, tt.isNonReporter, isNonReporter)
	}

	n.RevealRooks()
	assert.Equal(t, fmt.Sprintf("CHAT: %s, MESSAGE: <@userID1> is a rook in <#QWERTY123>! (Has enough worklogs: 27, enough commits: 2, and did not write standup!!!)\n<@userID2> is a rook in <#QWERTY123>! (Not enough worklogs: 3, enough commits: 30, and did not write standup!!!)\n<@userID3> is a rook in <#QWERTY123>! (Not enough worklogs: 0, no commits at all, and did not write standup!!!)\n<@userID4> is a rook in <#QWERTY123>! (Has enough worklogs: 13, enough commits: 20, and did not write standup!!!)\n", n.Config.ChanGeneral), ch.LastMessage)

	assert.NoError(t, n.DB.DeleteStandupUser(u1.SlackName, u1.ChannelID))
	assert.NoError(t, n.DB.DeleteStandupUser(u2.SlackName, u2.ChannelID))
	assert.NoError(t, n.DB.DeleteStandupUser(u3.SlackName, u3.ChannelID))
	assert.NoError(t, n.DB.DeleteStandupUser(u4.SlackName, u4.ChannelID))

	assert.NoError(t, n.DB.DeleteStandupTime(st.ChannelID))
}
