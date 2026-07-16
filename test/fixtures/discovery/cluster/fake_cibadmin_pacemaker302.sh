#!/usr/bin/env bash

# Pacemaker 3.0.2: both stonith-enabled (deprecated) and fencing-enabled (new) are present.
# CLIs write both names so checks reading either name still find a value.

cat <<'EOF'
<cib crm_feature_set="3.20.5" validate-with="pacemaker-4.0" epoch="6881" num_updates="0" admin_epoch="0" cib-last-written="Mon Nov 18 17:48:21 2019" update-origin="node01" update-client="crm_attribute" update-user="root" have-quorum="1" dc-uuid="1084783375">
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
    <resources>
      <clone id="mst_SAPHanaCon_PRD_HDB00">
        <meta_attributes id="mst_SAPHanaCon_PRD_HDB00-meta_attributes">
          <nvpair name="promotable" value="true" id="mst_SAPHanaCon_PRD_HDB00-meta_attributes-promotable"/>
          <nvpair name="clone-max" value="2" id="mst_SAPHanaCon_PRD_HDB00-meta_attributes-clone-max"/>
          <nvpair name="clone-node-max" value="1" id="mst_SAPHanaCon_PRD_HDB00-meta_attributes-clone-node-max"/>
        </meta_attributes>
        <primitive id="rsc_SAPHanaCon_PRD_HDB00" class="ocf" provider="suse" type="SAPHanaController">
          <instance_attributes id="rsc_SAPHanaCon_PRD_HDB00-instance_attributes">
            <nvpair name="SID" value="PRD" id="rsc_SAPHanaCon_PRD_HDB00-instance_attributes-SID"/>
            <nvpair name="InstanceNumber" value="00" id="rsc_SAPHanaCon_PRD_HDB00-instance_attributes-InstanceNumber"/>
          </instance_attributes>
          <operations>
            <op name="start" interval="0" timeout="3600" id="rsc_SAPHanaCon_PRD_HDB00-start-0"/>
            <op name="stop" interval="0" timeout="3600" id="rsc_SAPHanaCon_PRD_HDB00-stop-0"/>
            <op name="promote" interval="0" timeout="700" id="rsc_SAPHanaCon_PRD_HDB00-promote-0"/>
            <op name="monitor" interval="60" role="Promoted" timeout="700" id="rsc_SAPHanaCon_PRD_HDB00-monitor-60"/>
            <op name="monitor" interval="61" role="Unpromoted" timeout="700" id="rsc_SAPHanaCon_PRD_HDB00-monitor-61"/>
          </operations>
        </primitive>
      </clone>
      <clone id="cln_SAPHanaTop_PRD_HDB00">
        <meta_attributes id="cln_SAPHanaTop_PRD_HDB00-meta_attributes">
          <nvpair name="clone-node-max" value="1" id="cln_SAPHanaTop_PRD_HDB00-meta_attributes-clone-node-max"/>
          <nvpair name="interleave" value="true" id="cln_SAPHanaTop_PRD_HDB00-meta_attributes-interleave"/>
        </meta_attributes>
        <primitive id="rsc_SAPHanaTop_PRD_HDB00" class="ocf" provider="suse" type="SAPHanaTopology">
          <instance_attributes id="rsc_SAPHanaTop_PRD_HDB00-instance_attributes">
            <nvpair name="SID" value="PRD" id="rsc_SAPHanaTop_PRD_HDB00-instance_attributes-SID"/>
            <nvpair name="InstanceNumber" value="00" id="rsc_SAPHanaTop_PRD_HDB00-instance_attributes-InstanceNumber"/>
          </instance_attributes>
          <operations>
            <op name="start" interval="0" timeout="600" id="rsc_SAPHanaTop_PRD_HDB00-start-0"/>
            <op name="stop" interval="0" timeout="300" id="rsc_SAPHanaTop_PRD_HDB00-stop-0"/>
            <op name="monitor" interval="50" timeout="600" id="rsc_SAPHanaTop_PRD_HDB00-monitor-50"/>
          </operations>
        </primitive>
      </clone>
    </resources>
    <constraints/>
  </configuration>
  <status/>
</cib>
EOF
