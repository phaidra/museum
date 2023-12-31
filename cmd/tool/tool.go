package tool

import (
	"gopkg.in/yaml.v3"
	"museum/domain"
	"museum/ioc"
	"os"
)

func createToolContainer() *ioc.Container {
	c := ioc.NewContainer()
	ioc.RegisterSingleton[ApiClient](c, func() ApiClient {
		return &ApiClientImpl{
			BaseUrl: "http://localhost:8080",
		}
	})
	return c
}

func Create(filePath string) (*domain.Exhibit, string, error) {
	c := createToolContainer()

	_, err := os.Open(filePath)
	if err != nil {
		return nil, "", err
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", err
	}

	exhibit := &domain.Exhibit{}
	err = yaml.Unmarshal(content, exhibit)

	if err != nil {
		return nil, "", err
	}

	a := ioc.Get[ApiClient](c)
	id, err := a.CreateExhibit(exhibit)
	if err != nil {
		return nil, "", err
	}

	exhibit.Id = id

	return exhibit, a.GetBaseUrl() + "/exhibit/" + id, nil
}

func Delete(id string) error {
	c := createToolContainer()

	a := ioc.Get[ApiClient](c)
	err := a.DeleteExhibitById(id)
	if err != nil {
		return err
	}

	return nil
}

func Warmup(id string) (string, error) {
	c := createToolContainer()

	a := ioc.Get[ApiClient](c)
	exhibit, err := a.GetExhibitById(id)
	if err != nil {
		return "", err
	}

	event, err := domain.NewStartEvent(exhibit.ToExhibit())
	if err != nil {
		return "", err
	}

	err = a.CreateEvent(&event)
	if err != nil {
		return "", err
	}

	return a.GetBaseUrl() + "/exhibit/" + id, nil
}

func List() (string, []domain.ExhibitDto, error) {
	c := createToolContainer()

	a := ioc.Get[ApiClient](c)
	exhibits, err := a.GetAllExhibits()
	if err != nil {
		return "", nil, err
	}

	return a.GetBaseUrl(), exhibits, nil
}
