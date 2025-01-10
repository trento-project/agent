package gatherers_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/test/helpers"
)

type SapProfilesTestSuite struct {
	suite.Suite
}

func TestSapProfilesTestSuite(t *testing.T) {
	suite.Run(t, new(SapProfilesTestSuite))
}

func (suite *SapProfilesTestSuite) TestSapProfilesSuccess() {
	appFS := afero.NewMemMapFs()

	err := appFS.MkdirAll("/usr/sap/PRD", 0644)
	suite.NoError(err)
	err = appFS.MkdirAll("/usr/sap/QAS", 0644)
	suite.NoError(err)

	defaultProfileFile, _ := os.Open(helpers.GetFixturePath("gatherers/sap_profile.default"))
	defaultProfileContent, _ := io.ReadAll(defaultProfileFile)
	err = afero.WriteFile(appFS, "/sapmnt/PRD/profile/DEFAULT.PFL", defaultProfileContent, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/PRD/profile/DEFAULT.1.PFL", []byte{}, 0644)
	suite.NoError(err)

	ascsProfileFile, _ := os.Open(helpers.GetFixturePath("gatherers/sap_profile.ascs"))
	ascsProfileConcent, _ := io.ReadAll(ascsProfileFile)
	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/QAS_ASCS00_sapqasas", ascsProfileConcent, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/QAS_ASCS00_sapqasas.1", []byte{}, 0644)
	suite.NoError(err)
	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/QAS_ASCS00_sapqasas.bak", []byte{}, 0644)
	suite.NoError(err)

	minimalProfileFile, _ := os.Open(helpers.GetFixturePath("gatherers/sap_profile.minimal"))
	minimalProfileContent, _ := io.ReadAll(minimalProfileFile)
	err = afero.WriteFile(appFS, "/sapmnt/QAS/profile/DEFAULT.PFL", minimalProfileContent, 0644)
	suite.NoError(err)

	gatherer := gatherers.NewSapProfilesGatherer(appFS)

	fr := []entities.FactRequest{{
		Name:     "sap_profiles",
		Gatherer: "sap_profiles",
		CheckID:  "check1",
	}}

	expectedFacts := []entities.Fact{
		{
			Name:    "sap_profiles",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"PRD": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"profiles": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"name": &entities.FactValueString{
												Value: "DEFAULT.PFL",
											},
											"path": &entities.FactValueString{
												Value: "/sapmnt/PRD/profile/DEFAULT.PFL",
											},
											"content": &entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"SAPDBHOST": &entities.FactValueString{
														Value: "192.168.140.12",
													},
													"SAPGLOBALHOST": &entities.FactValueString{
														Value: "sapha1as",
													},
													"SAPSYSTEMNAME": &entities.FactValueString{
														Value: "HA1",
													},
													"dbs/hdb/dbname": &entities.FactValueString{
														Value: "PRD",
													},
													"dbs/hdb/schema": &entities.FactValueString{
														Value: "SAPABAP1",
													},
													"enque/deque_wait_answer": &entities.FactValueBool{
														Value: true,
													},
													"enque/process_location": &entities.FactValueString{
														Value: "REMOTESA",
													},
													"enque/serverhost": &entities.FactValueString{
														Value: "sapha1as",
													},
													"enque/serverinst": &entities.FactValueInt{
														Value: 0,
													},
													"gw/acl_mode": &entities.FactValueInt{
														Value: 1,
													},
													"gw/sec_info": &entities.FactValueString{
														Value: "$(DIR_GLOBAL)$(DIR_SEP)secinfo$(FT_DAT)",
													},
													"icf/user_recheck": &entities.FactValueInt{
														Value: 1,
													},
													"icm/HTTP/ASJava/disable_url_session_tracking": &entities.FactValueBool{
														Value: true,
													},
													"is/HTTP/show_detailed_errors": &entities.FactValueBool{
														Value: false,
													},
													"j2ee/dbhost": &entities.FactValueString{
														Value: "192.168.140.12",
													},
													"j2ee/dbname": &entities.FactValueString{
														Value: "PRD",
													},
													"j2ee/dbtype": &entities.FactValueString{
														Value: "hdb",
													},
													"login/password_downwards_compatibility": &entities.FactValueInt{
														Value: 0,
													},
													"login/system_client": &entities.FactValueInt{
														Value: 1,
													},
													"rdisp/mshost": &entities.FactValueString{
														Value: "sapha1as",
													},
													"rdisp/msserv": &entities.FactValueString{
														Value: "sapmsHA1",
													},
													"rdisp/msserv_internal": &entities.FactValueInt{
														Value: 3900,
													},
													"rsdb/ssfs_connect": &entities.FactValueInt{
														Value: 0,
													},
													"rsec/ssfs_datapath": &entities.FactValueString{
														Value: "$(DIR_GLOBAL)$(DIR_SEP)security$(DIR_SEP)rsecssfs$(DIR_SEP)data",
													},
													"rsec/ssfs_keypath": &entities.FactValueString{
														Value: "$(DIR_GLOBAL)$(DIR_SEP)security$(DIR_SEP)rsecssfs$(DIR_SEP)key",
													},
													"service/protectedwebmethods": &entities.FactValueString{
														Value: "SDEFAULT",
													},
													"system/type": &entities.FactValueString{
														Value: "ABAP",
													},
													"vmcj/enable": &entities.FactValueString{
														Value: "off",
													},
												},
											},
										},
									},
								},
							},
						},
					},
					"QAS": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"profiles": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"name": &entities.FactValueString{
												Value: "DEFAULT.PFL",
											},
											"path": &entities.FactValueString{
												Value: "/sapmnt/QAS/profile/DEFAULT.PFL",
											},
											"content": &entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"SAPSYSTEMNAME": &entities.FactValueString{
														Value: "QAS",
													},
													"SAPGLOBALHOST": &entities.FactValueString{
														Value: "sapha1as",
													},
												},
											},
										},
									},
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"name": &entities.FactValueString{
												Value: "QAS_ASCS00_sapqasas",
											},
											"path": &entities.FactValueString{
												Value: "/sapmnt/QAS/profile/QAS_ASCS00_sapqasas",
											},
											"content": &entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"DIR_CT_RUN": &entities.FactValueString{
														Value: "$(DIR_EXE_ROOT)$(DIR_SEP)$(OS_UNICODE)$(DIR_SEP)linuxx86_64",
													},
													"DIR_EXECUTABLE": &entities.FactValueString{
														Value: "$(DIR_INSTANCE)/exe",
													},
													"DIR_PROFILE": &entities.FactValueString{
														Value: "$(DIR_INSTALL)$(DIR_SEP)profile",
													},
													"Execute_00": &entities.FactValueString{
														Value: "immediate $(DIR_CT_RUN)/sapcpe$(FT_EXE) pf=$(_PF) $(_CPARG0)",
													},
													"Execute_01": &entities.FactValueString{
														Value: "immediate $(DIR_CT_RUN)/sapcpe$(FT_EXE) pf=$(_PF) $(_CPARG1)",
													},
													"Execute_02": &entities.FactValueString{
														Value: "local rm -f $(_MS)",
													},
													"Execute_03": &entities.FactValueString{
														Value: "local ln -s -f $(DIR_EXECUTABLE)/msg_server$(FT_EXE) $(_MS)",
													},
													"Execute_04": &entities.FactValueString{
														Value: "local rm -f $(_EN)",
													},
													"Execute_05": &entities.FactValueString{
														Value: "local ln -s -f $(DIR_EXECUTABLE)/enserver$(FT_EXE) $(_EN)",
													},
													"INSTANCE_NAME": &entities.FactValueString{
														Value: "ASCS00",
													},
													"Restart_Program_00": &entities.FactValueString{
														Value: "local $(_MS) pf=$(_PF)",
													},
													"SAPLOCALHOST": &entities.FactValueString{
														Value: "sapnwpas",
													},
													"SAPSYSTEM": &entities.FactValueInt{
														Value: 0,
													},
													"SAPSYSTEMNAME": &entities.FactValueString{
														Value: "NWP",
													},
													"SETENV_00": &entities.FactValueString{
														Value: "DIR_LIBRARY=$(DIR_LIBRARY)",
													},
													"SETENV_01": &entities.FactValueString{
														Value: "LD_LIBRARY_PATH=$(DIR_LIBRARY):%(LD_LIBRARY_PATH)",
													},
													"SETENV_02": &entities.FactValueString{
														Value: "SHLIB_PATH=$(DIR_LIBRARY):%(SHLIB_PATH)",
													},
													"SETENV_03": &entities.FactValueString{
														Value: "LIBPATH=$(DIR_LIBRARY):%(LIBPATH)",
													},
													"SETENV_04": &entities.FactValueString{
														Value: "PATH=$(DIR_EXECUTABLE):%(PATH)",
													},
													"SETENV_05": &entities.FactValueString{
														Value: "SECUDIR=$(DIR_INSTANCE)/sec",
													},
													"Start_Program_01": &entities.FactValueString{
														Value: "local $(_EN) pf=$(_PF)",
													},
													"_CPARG0": &entities.FactValueString{
														Value: "list:$(DIR_CT_RUN)/scs.lst",
													},
													"_CPARG1": &entities.FactValueString{
														Value: "list:$(DIR_CT_RUN)/sapcrypto.lst",
													},
													"_EN": &entities.FactValueString{
														Value: "en.sap$(SAPSYSTEMNAME)_$(INSTANCE_NAME)",
													},
													"_MS": &entities.FactValueString{
														Value: "ms.sap$(SAPSYSTEMNAME)_$(INSTANCE_NAME)",
													},
													"_PF": &entities.FactValueString{
														Value: "$(DIR_PROFILE)/NWP_ASCS00_sapnwpas",
													},
													"enque/encni/set_so_keepalive": &entities.FactValueBool{
														Value: true,
													},
													"enque/server/max_query_requests": &entities.FactValueInt{
														Value: 5000,
													},
													"enque/server/max_requests": &entities.FactValueInt{
														Value: 5000,
													},
													"enque/server/replication": &entities.FactValueBool{
														Value: true,
													},
													"enque/server/threadcount": &entities.FactValueInt{
														Value: 4,
													},
													"enque/snapshot_pck_ids": &entities.FactValueInt{
														Value: 1600,
													},
													"enque/table_size": &entities.FactValueInt{
														Value: 64000,
													},
													"ms/server_port_0": &entities.FactValueString{
														Value: "PROT=HTTP,PORT=81$$",
													},
													"ms/standalone": &entities.FactValueInt{
														Value: 1,
													},
													"rdisp/enqname": &entities.FactValueString{
														Value: "$(rdisp/myname)",
													},
													"service/halib": &entities.FactValueString{
														Value: "$(DIR_CT_RUN)/saphascriptco.so",
													},
													"service/halib_cluster_connector": &entities.FactValueString{
														Value: "/usr/bin/sap_suse_cluster_connector",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	results, err := gatherer.Gather(context.Background(), fr)
	suite.NoError(err)
	suite.EqualValues(expectedFacts, results)
}

func (suite *SapProfilesTestSuite) TestSapProfilesNoProfiles() {
	appFS := afero.NewMemMapFs()

	err := appFS.MkdirAll("/usr/sap/PRD", 0644)
	suite.NoError(err)

	err = appFS.MkdirAll("/sapmnt/PRD/profile", 0755)
	suite.NoError(err)

	gatherer := gatherers.NewSapProfilesGatherer(appFS)

	fr := []entities.FactRequest{{
		Name:     "sap_profiles",
		Gatherer: "sap_profiles",
		CheckID:  "check1",
	}}

	expectedFacts := []entities.Fact{
		{
			Name:    "sap_profiles",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"PRD": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"profiles": &entities.FactValueList{
								Value: []entities.FactValue{},
							},
						},
					},
				},
			},
		},
	}

	results, err := gatherer.Gather(context.Background(), fr)
	suite.NoError(err)
	suite.EqualValues(expectedFacts, results)
}

func (suite *SapProfilesTestSuite) TestSapProfilesInvalidProfile() {
	appFS := afero.NewMemMapFs()

	err := appFS.MkdirAll("/usr/sap/PRD", 0644)
	suite.NoError(err)

	err = afero.WriteFile(appFS, "/sapmnt/PRD/profile/DEFAULT.PFL", []byte("invalid"), 0644)
	suite.NoError(err)

	gatherer := gatherers.NewSapProfilesGatherer(appFS)

	fr := []entities.FactRequest{{
		Name:     "sap_profiles",
		Gatherer: "sap_profiles",
		CheckID:  "check1",
	}}

	result, err := gatherer.Gather(context.Background(), fr)
	suite.Nil(result)
	suite.EqualError(err, "fact gathering error: sap-profiles-file-system-error - "+
		"error reading the sap profiles file system: could not parse profile file: error "+
		"on line 1: missing =")
}
