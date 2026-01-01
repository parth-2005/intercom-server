package main

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var fcmClient *messaging.Client

// InitFirebase initializes the Firebase Admin SDK
func InitFirebase(serviceAccountPath string) error {
	opt := option.WithCredentialsFile(serviceAccountPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return err
	}

	fcmClient = client
	log.Println("Firebase Admin SDK initialized")
	return nil
}

// sendPushNotification sends a data-only message for flutter_callkit_incoming
func sendPushNotification(fcmToken string, payload map[string]string) error {
	if fcmClient == nil {
		log.Println("FCM client not initialized")
		return nil
	}

	message := &messaging.Message{
		Token: fcmToken,
		Data:  payload,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					ContentAvailable: true,
				},
			},
		},
	}

	response, err := fcmClient.Send(context.Background(), message)
	if err != nil {
		log.Printf("Failed to send FCM message: %v", err)
		return err
	}

	log.Printf("Successfully sent FCM message: %s", response)
	return nil
}
