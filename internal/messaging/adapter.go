// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package messaging

type Adapter interface {
	Unsubscribe() error
	Listen(handle func(contentType string, message []byte) error) error
	Publish(routingKey, contentType string, message []byte) error
}
