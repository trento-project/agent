// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package messaging

type Adapter interface {
	Unsubscribe() error
	Listen(handle func(contentType string, message []byte) error) error
	Publish(routingKey, contentType string, message []byte) error
}
