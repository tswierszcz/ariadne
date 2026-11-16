package main

import (
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/profile"
)

// TestMessageListenerRoots: the spring profile's roots include message
// listener methods — @JmsListener, @KafkaListener, @RabbitListener (also
// repeated: their @Repeatable containers), and the public methods of a
// class-level @KafkaListener/@RabbitListener (its @KafkaHandler/
// @RabbitHandler methods). testdata/listeners holds name-only stand-ins
// for the annotations (the profile matches roots by name); the real
// container dispatch is checked by scripts/prepare-listeners-demo.sh.
func TestMessageListenerRoots(t *testing.T) {
	h, _, err := hierarchy.Build("../../testdata/listeners/app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	cat, err := profile.LoadDefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}
	prof, _ := cat.Get("spring")
	g := callgraph.Build(h, nil)
	var got []string
	for _, id := range prof.AppRoots(g, h) {
		if k := g.Key(id); strings.HasPrefix(k, "fx.msg.") {
			got = append(got, k)
		}
	}
	sort.Strings(got)
	// A class-level annotation also makes the public constructor a root
	// (the existing class-annotation rule; the container creates the bean
	// anyway).
	want := []string{
		"fx.msg.KafkaClassListener#<init>()V",
		"fx.msg.KafkaClassListener#onOther(Ljava/lang/Object;)V",
		"fx.msg.KafkaClassListener#onString(Ljava/lang/String;)V",
		"fx.msg.Listeners#onJms(Ljava/lang/String;)V",
		"fx.msg.Listeners#onKafka(Ljava/lang/String;)V",
		"fx.msg.Listeners#onRabbit(Ljava/lang/String;)V",
		"fx.msg.Listeners#onTwoQueues(Ljava/lang/String;)V",
		"fx.msg.Listeners#onTwoTopics(Ljava/lang/String;)V",
		"fx.msg.RabbitClassListener#<init>()V",
		"fx.msg.RabbitClassListener#onString(Ljava/lang/String;)V",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("listener roots:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
