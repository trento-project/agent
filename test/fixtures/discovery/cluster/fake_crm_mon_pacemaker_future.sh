#!/usr/bin/env bash

cat <<'EOF'
<?xml version="1.0"?>
<crm_mon version="4.0.0">
  <summary>
    <nodes_configured number="2"/>
    <resources_configured number="1" disabled="0" blocked="0"/>
    <last_change time="Fri Oct 18 11:48:22 2019"/>
    <cluster_options fencing-enabled="true"/>
  </summary>
  <nodes/>
  <node_attributes/>
  <node_history/>
  <resources>
    <resource id="rsc-1" resource_agent="ocf::heartbeat:Dummy" role="Started"
              active="true" removed="false" blocked="false"
              managed="true" failed="false" failure_ignored="false" nodes_running_on="1"/>
  </resources>
</crm_mon>
EOF
