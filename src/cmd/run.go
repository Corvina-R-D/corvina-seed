package cmd

import (
	"context"
	"corvina/corvina-seed/src/seed"
	"corvina/corvina-seed/src/seed/dto"
	"errors"

	"github.com/manifoldco/promptui"
)

type CtxKey string

const ApiKey CtxKey = "apiKey"

func Run(ctx context.Context) error {

	apiKey, err := takeApiKeyFromCtxOrAskIt(ctx)
	if err != nil {
		return err
	}

	executeInput := dto.ExecuteInDTO{
		ApiKey: apiKey,
	}
	err = seed.Execute(ctx, executeInput)
	if err != nil {
		return err
	}

	return nil
}

func takeApiKeyFromCtxOrAskIt(ctx context.Context) (string, error) {
	apiKey := ctx.Value(ApiKey)

	if apiKey == nil || apiKey == "" {
		apiKey, err := askApiKey()
		if err != nil {
			return "", err
		}

		return apiKey, nil
	}

	err := validateApiKey(apiKey.(string))
	if err != nil {
		return "", err
	}

	return apiKey.(string), nil
}

func askApiKey() (string, error) {
	prompt := promptui.Prompt{
		Label:    "API Key",
		Validate: validateApiKey,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

func validateApiKey(apiKey string) error {
	if len(apiKey) < 1 {
		return errors.New("assign an api key with a minimum of 1 character")
	}

	return nil
}
