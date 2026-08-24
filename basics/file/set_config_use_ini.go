package main

import (
	"gopkg.in/ini.v1"
	"log"
)

type ConfigTopic struct {
	file    string
	section string
	key     string
	value   string
}

var configs []ConfigTopic

func init() {
	configs = append(configs, ConfigTopic{file: "/etc/kolla/nova-api/nova.conf", section: "oslo_messaging_notifications", key: "topics", value: "nova_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/nova-api/nova.conf", section: "DEFAULT", key: "versioned_notifications_topics", value: "nova_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/nova-scheduler/nova.conf", section: "oslo_messaging_notifications", key: "topics", value: "nova_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/nova-novncproxy/nova.conf", section: "oslo_messaging_notifications", key: "topics", value: "nova_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/nova-compute/nova.conf", section: "oslo_messaging_notifications", key: "topics", value: "nova_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/nova-conductor/nova.conf", section: "oslo_messaging_notifications", key: "topics", value: "nova_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/nova-serialproxy/nova.conf", section: "oslo_messaging_notifications", key: "topics", value: "nova_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/nova-compute-ironic/nova.conf", section: "oslo_messaging_notifications", key: "topics", value: "nova_ecs_voneyun_topic"})

	configs = append(configs, ConfigTopic{file: "/etc/kolla/ironic-conductor/ironic.conf", section: "DEFAULT", key: "versioned_notifications_topics", value: "nova_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/ironic-api/ironic.conf", section: "DEFAULT", key: "versioned_notifications_topics", value: "nova_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/ironic-neutron-agent/neutron.conf", section: "oslo_messaging_notifications", key: "topics", value: "neutron_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/nova-compute-ironic/nova.conf", section: "oslo_messaging_notifications", key: "topics", value: "nova_ecs_voneyun_topic"})

	configs = append(configs, ConfigTopic{file: "/etc/kolla/cinder-backup/cinder.conf", section: "oslo_messaging_notifications", key: "topics", value: "cinder_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/cinder-api/cinder.conf", section: "oslo_messaging_notifications", key: "topics", value: "cinder_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/cinder-volume/cinder.conf", section: "oslo_messaging_notifications", key: "topics", value: "cinder_ecs_voneyun_topic"})
	configs = append(configs, ConfigTopic{file: "/etc/kolla/cinder-scheduler/cinder.conf", section: "oslo_messaging_notifications", key: "topics", value: "cinder_ecs_voneyun_topic"})
}

func setConfigTopic(config ConfigTopic) {
	var err error
	cfg, err := ini.Load(config.file)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	val, err := cfg.GetSection(config.section)
	if err != nil {
		log.Fatalf("Failed to get section %v", config.section)
	}

	val.Key(config.key).SetValue(config.value)
	err = cfg.SaveTo(config.file)
	if err != nil {
		log.Fatalf("Failed to save: %v", err)
	}
}

func main() {
	for _, config := range configs {
		setConfigTopic(config)
	}
}
