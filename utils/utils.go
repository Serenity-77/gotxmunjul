package libs

import (
    "os"
    "errors"
    "reflect"
)


var (
    osStatFunc = os.Stat
)

func TypeNameOf(obj interface{}) string {
    t := reflect.TypeOf(obj)

    if t.Kind() == reflect.Ptr {
        return t.Elem().Name()
    }

    return t.Name()
}

func FileExists(path string) (bool, error) {
    var err error

    _, err = osStatFunc(path)

    if err != nil && errors.Is(err, os.ErrNotExist) {
        return false, nil
    } else if err != nil {
        return false, err
    }

    return true, nil
}
