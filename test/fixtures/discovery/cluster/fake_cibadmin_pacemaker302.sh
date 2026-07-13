#!/usr/bin/env bash

# Pacemaker 3.0.2: both stonith-enabled (deprecated) and fencing-enabled (new) are present.
# CLIs write both names so checks reading either name still find a value.

cat <<'EOF'
<cib crm_feature_set="3.20.5" validate-with="pacemaker-3.2" epoch="6881" num_updates="0" admin_epoch="0" cib-last-written="Mon Nov 18 17:48:21 2019" update-origin="node01" update-client="crm_attribute" update-user="root" have-quorum="1" dc-uuid="1084783375">
  <configuration>
    <crm_config>
      <cluster_property_set id="cib-bootstrap-options">
        <nvpair id="cib-bootstrap-options-have-watchdog" name="have-watchdog" value="true"/>
        <nvpair id="cib-bootstrap-options-dc-version" name="dc-version" value="3.0.2"/>
        <nvpair id="cib-bootstrap-options-cluster-infrastructure" name="cluster-infrastructure" value="corosync"/>
        <nvpair id="cib-bootstrap-options-cluster-name" name="cluster-name" value="hana_cluster"/>
        <nvpair name="stonith-enabled" value="true" id="cib-bootstrap-options-stonith-enabled"/>
        <nvpair name="fencing-enabled" value="true" id="cib-bootstrap-options-fencing-enabled"/>
        <nvpair name="placement-strategy" value="balanced" id="cib-bootstrap-options-placement-strategy"/>
      </cluster_property_set>
    </crm_config>
    <nodes/>
    <resources/>
    <constraints/>
  </configuration>
  <status/>
</cib>
EOF
