package imexport

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"rxdrag.com/entify/app"
	"rxdrag.com/entify/consts"
	"rxdrag.com/entify/scalars"
	"rxdrag.com/entify/storage"
	"rxdrag.com/entify/utils"
)

func (m *ImExportModule) MutationFields() []*graphql.Field {
	if !app.Installed {
		return []*graphql.Field{}
	}
	return []*graphql.Field{
		{
			Name: IMPORT_APP,
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				ARG_APP_FILE: &graphql.ArgumentConfig{
					Type: scalars.UploadType,
				},
				ARG_APP_ID: &graphql.ArgumentConfig{
					Type: graphql.ID,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				defer utils.PrintErrorStack()
				return m.importResolve(p)
			},
		},
	}
}

func (m *ImExportModule) importResolve(p graphql.ResolveParams) (interface{}, error) {
	appId := m.app.AppId
	if p.Args[ARG_APP_ID] != nil {
		intId, err := strconv.ParseUint(p.Args[ARG_APP_ID].(string), 10, 64)
		if err != nil {
			log.Panic(err.Error())
		}
		appId = intId
	}
	upload := p.Args[ARG_APP_FILE].(storage.File)
	fileInfo := upload.Save(TEMP_DATAS)

	r, err := zip.OpenReader(consts.STATIC_PATH + "/" + fileInfo.Path)
	if err != nil {
		log.Panic(err.Error())
	}

	var appJsonFile *zip.File
	for _, f := range r.File {
		if f.Name == APP_JON {
			appJsonFile = f
		}
	}

	if appJsonFile == nil {
		log.Panic(fmt.Sprintf("Can not find %s in upload file", APP_JON))
	}

	appMap := readAppJsonFile(appJsonFile)

	if appMap["plugins"] != nil {
		plugins := appMap["plugins"].([]map[string]interface{})
		for _, plugin := range plugins {
			if plugin["type"] != "debug" {
				pluginFiles := getPluginFiles(plugin["url"].(string), r.File)
				hostPath := getHostPath(p.Context)
				pluginName := uuid.New().String()
				relativePath := fmt.Sprintf("%s/app%d/plugins/%s", consts.STATIC_PATH, appId, pluginName)
				plugin["url"] = hostPath + relativePath
				for i := range pluginFiles {
					extractAndWriteFile(relativePath, pluginFiles[i])
				}
			}
		}
	}
	return true, nil
}

func getPluginFiles(pluginPath string, arr []*zip.File) []*zip.File {
	files := []*zip.File{}
	for i := range arr {
		if strings.Index(arr[i].Name, fmt.Sprintf("plugins/%s/", pluginPath)) == 0 {
			files = append(files, arr[i])
		}
	}
	return files
}

func readAppJsonFile(f *zip.File) map[string]interface{} {
	rc, err := f.Open()
	if err != nil {
		log.Panic(err.Error())
	}
	defer func() {
		if err := rc.Close(); err != nil {
			log.Panic(err.Error())
		}
	}()

	jsonByte := []byte{}
	length, err := rc.Read(jsonByte)

	if err != nil {
		log.Panic(err.Error())
	}

	if length == 0 {
		log.Panic("app.json is emperty")
	}

	appMap := map[string]interface{}{}
	err = json.Unmarshal(jsonByte, &appMap)
	if err != nil {
		log.Panic(err.Error())
	}

	return appMap
}

func extractAndWriteFile(destination string, f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := rc.Close(); err != nil {
			panic(err)
		}
	}()

	path := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(path, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("%s: illegal file path", path)
	}

	if f.FileInfo().IsDir() {
		err = os.MkdirAll(path, f.Mode())
		if err != nil {
			return err
		}
	} else {
		err = os.MkdirAll(filepath.Dir(path), f.Mode())
		if err != nil {
			return err
		}

		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()

		_, err = io.Copy(f, rc)
		if err != nil {
			return err
		}
	}

	return nil
}
