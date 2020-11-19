// +build integration

/*
Copyright 2020 Comcast Cable Communications Management, LLC
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/vinyldns/go-vinyldns/vinyldns"
)

var _ = Describe("its commands for working with zones", func() {
	var (
		session   *gexec.Session
		err       error
		args      []string
		zonesArgs []string
		group     *vinyldns.Group
		makeGroup = func() *vinyldns.Group {
			return &vinyldns.Group{
				Name:        "zones-test-group",
				Description: "description",
				Email:       "email@email.com",
				Admins: []vinyldns.User{{
					UserName: "ok",
					ID:       "ok",
				}},
				Members: []vinyldns.User{{
					UserName: "ok",
					ID:       "ok",
				}},
			}
		}
		makeZone = func(name, adminGroupID string) *vinyldns.Zone {
			return &vinyldns.Zone{
				Name:         name,
				Email:        "email@email.com",
				AdminGroupID: adminGroupID,
			}
		}
		cleanUp = func(group *vinyldns.Group, name string, deleteZones bool) {
			var zones []vinyldns.Zone
			var id string

			for {
				if !deleteZones {
					break
				}

				zones, err = vinylClient.Zones()
				Expect(err).NotTo(HaveOccurred())

				if len(zones) != 0 {
					break
				}
			}

			for _, z := range zones {
				if !deleteZones {
					break
				}

				if z.Name == name {
					id = z.ID
					_, err = vinylClient.ZoneDelete(id)
					Expect(err).NotTo(HaveOccurred())
					break
				}
			}

			for {
				if !deleteZones {
					break
				}

				var exists bool
				exists, err = vinylClient.ZoneExists(id)
				Expect(err).NotTo(HaveOccurred())

				if !exists {
					break
				}
			}

			// There's a window of time following zone deletion in which
			// VinylDNS continues to believe the group is a zone admin.
			// We sleep for 3 seconds to allow VinylDNS to get itself straight.
			time.Sleep(3 * time.Second)

			_, err = vinylClient.GroupDelete(group.ID)
			Expect(err).NotTo(HaveOccurred())

			for {
				groups, err := vinylClient.Groups()
				Expect(err).NotTo(HaveOccurred())

				if len(groups) == 0 {
					break
				}
			}
		}
	)

	JustBeforeEach(func() {
		args = append(baseArgs, zonesArgs...)
		cmd := exec.Command(exe, args...)
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	})

	JustAfterEach(func() {
		session.Terminate().Wait()
	})

	Describe("its 'zones' command", func() {
		Context("when it's passed '--help'", func() {
			BeforeEach(func() {
				zonesArgs = []string{
					"zones",
					"--help",
				}
			})

			It("prints a useful description", func() {
				Eventually(session.Out, 5).Should(gbytes.Say("List all vinyldns zones"))
			})
		})

		Context("when no zones exist", func() {
			Context("when not passed an --output", func() {
				BeforeEach(func() {
					zonesArgs = []string{
						"zones",
					}
				})

				It("prints the correct data", func() {
					Eventually(session.Out, 5).Should(gbytes.Say("No zones found"))
				})
			})

			Context("when passed an --output=json", func() {
				BeforeEach(func() {
					zonesArgs = []string{
						"--output=json",
						"zones",
					}
				})

				It("prints the correct data", func() {
					Eventually(session.Out, 5).Should(gbytes.Say(`\[\]`))
				})
			})
		})

		Context("when zones exist", func() {
			var (
				zone *vinyldns.ZoneUpdateResponse
				name string = "vinyldns."
			)

			BeforeEach(func() {
				group, err = vinylClient.GroupCreate(makeGroup())
				Expect(err).NotTo(HaveOccurred())

				zone, err = vinylClient.ZoneCreate(makeZone(name, group.ID))
				Expect(err).NotTo(HaveOccurred())

				// wait to be sure the zone is fully created
				// TODO: this can be improved
				time.Sleep(3 * time.Second)
			})

			AfterEach(func() {
				_, err = vinylClient.ZoneDelete(zone.Zone.ID)
				Expect(err).NotTo(HaveOccurred())

				for {
					exists, err := vinylClient.ZoneExists(zone.Zone.ID)
					Expect(err).NotTo(HaveOccurred())

					if !exists {
						break
					}
				}

				_, err = vinylClient.GroupDelete(group.ID)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when it's not passed the --output=json option", func() {
				BeforeEach(func() {
					zonesArgs = []string{
						"zones",
					}
				})

				It("prints zone details", func() {
					output := fmt.Sprintf(`+-----------+--------------------------------------+
|   NAME    |                  ID                  |
+-----------+--------------------------------------+
| vinyldns. | %s |
+-----------+--------------------------------------+`, zone.Zone.ID)

					Eventually(func() string {
						return string(session.Out.Contents())
					}).Should(ContainSubstring(output))
				})
			})
		})
	})

	Describe("its 'zone' command", func() {
		Context("when it's passed '--help'", func() {
			BeforeEach(func() {
				zonesArgs = []string{
					"zone",
					"--help",
				}
			})

			It("prints a useful description", func() {
				Eventually(session.Out, 5).Should(gbytes.Say("view zone details"))
			})
		})

		Context("when the zone exists", func() {
			var (
				zone *vinyldns.ZoneUpdateResponse
				name string = "vinyldns."
			)

			BeforeEach(func() {
				group, err = vinylClient.GroupCreate(makeGroup())
				Expect(err).NotTo(HaveOccurred())

				zone, err = vinylClient.ZoneCreate(makeZone(name, group.ID))
				Expect(err).NotTo(HaveOccurred())

				// wait to be sure the zone is fully created
				time.Sleep(3 * time.Second)
			})

			AfterEach(func() {
				_, err = vinylClient.ZoneDelete(zone.Zone.ID)
				Expect(err).NotTo(HaveOccurred())

				for {
					exists, err := vinylClient.ZoneExists(zone.Zone.ID)
					Expect(err).NotTo(HaveOccurred())

					if !exists {
						break
					}
				}

				_, err = vinylClient.GroupDelete(group.ID)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("it's passed a '--zone-name'", func() {
				BeforeEach(func() {
					zonesArgs = []string{
						"zone",
						fmt.Sprintf("--zone-name=%s", name),
					}
				})

				It("prints the zone's details", func() {
					output := fmt.Sprintf(`+--------+--------------------------------------+
| Name   | %s                            |
+--------+--------------------------------------+
| ID     | %s |
+--------+--------------------------------------+
| Status | Active                               |
+--------+--------------------------------------+`, name, zone.Zone.ID)

					Eventually(func() string {
						return string(session.Out.Contents())
					}).Should(ContainSubstring(output))

				})
			})

			Context("it's passed a '--zone-id'", func() {
				BeforeEach(func() {
					zonesArgs = []string{
						"zone",
						fmt.Sprintf("--zone-id=%s", zone.Zone.ID),
					}
				})

				It("prints the zone's details", func() {
					output := fmt.Sprintf(`+--------+--------------------------------------+
| Name   | %s                            |
+--------+--------------------------------------+
| ID     | %s |
+--------+--------------------------------------+
| Status | Active                               |
+--------+--------------------------------------+`, name, zone.Zone.ID)

					Eventually(func() string {
						return string(session.Out.Contents())
					}).Should(ContainSubstring(output))
				})
			})
		})
	})

	Describe("its 'zone-create' command", func() {
		Context("when it's passed '--help'", func() {
			BeforeEach(func() {
				zonesArgs = []string{
					"zone-create",
					"--help",
				}
			})

			It("prints a useful description", func() {
				Eventually(session.Out, 5).Should(gbytes.Say("Create a zone"))
			})
		})

		Context("when it's not passed connection details", func() {
			var (
				name string = "vinyldns."
			)

			BeforeEach(func() {
				group, err = vinylClient.GroupCreate(makeGroup())
				Expect(err).NotTo(HaveOccurred())

				zonesArgs = []string{
					"zone-create",
					fmt.Sprintf("--name=%s", name),
					"--email=admin@test.com",
					fmt.Sprintf("--admin-group-name=%s", group.Name),
				}
			})

			AfterEach(func() {
				cleanUp(group, name, true)
			})

			It("prints a message reporting that the zone has been created", func() {
				Eventually(session.Out, 5).Should(gbytes.Say("Created zone vinyldns."))
			})
		})

		Context("when it's passed valid connection details", func() {
			var (
				name string = "vinyldns."
			)

			BeforeEach(func() {
				group, err = vinylClient.GroupCreate(makeGroup())
				Expect(err).NotTo(HaveOccurred())

				zonesArgs = []string{
					"zone-create",
					fmt.Sprintf("--name=%s", name),
					"--email=admin@test.com",
					fmt.Sprintf("--admin-group-name=%s", group.Name),
					"--zone-connection-key-name=vinyldns.",
					"--zone-connection-key=nzisn+4G2ldMn0q1CV3vsg==",
					"--zone-connection-primary-server=vinyldns-bind9",
					fmt.Sprintf("--transfer-connection-key-name=%s", name),
					"--transfer-connection-key=nzisn+4G2ldMn0q1CV3vsg==",
					"--transfer-connection-primary-server=vinyldns-bind9",
				}
			})

			AfterEach(func() {
				cleanUp(group, name, true)
			})

			It("prints a message reporting that the zone has been created", func() {
				Eventually(session.Out, 5).Should(gbytes.Say("Created zone vinyldns."))
			})
		})

		Context("when it's passed invalid connection details", func() {
			var (
				name string = "vinyldns."
			)

			BeforeEach(func() {
				group, err = vinylClient.GroupCreate(makeGroup())
				Expect(err).NotTo(HaveOccurred())

				zonesArgs = []string{
					"zone-create",
					fmt.Sprintf("--name=%s", name),
					"--email=admin@test.com",
					fmt.Sprintf("--admin-group-name=%s", group.Name),
					"--zone-connection-key=nzisn+4G2ldMn0q1CV3vsg==",
					"--zone-connection-primary-server=vinyldns-bind9",
				}
			})

			AfterEach(func() {
				cleanUp(group, name, false)
			})

			It("prints an explanatory message to stderr", func() {
				Eventually(session.Err, 5).Should(gbytes.Say("zone connection requires '--zone-connection-key-name', '--zone-connection-key', and '--zone-connection-primary-server'"))
			})

			It("exits 1", func() {
				Eventually(session, 3).Should(gexec.Exit(1))
			})
		})
	})
})
