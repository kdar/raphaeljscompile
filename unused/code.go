type Undefined struct{}

type Object struct {
  _properties map[string]interface{}
}

func NewObject(engine *v8.Engine) (*v8.ObjectTemplate, *v8.FunctionTemplate) {
  ftConstructor := engine.NewFunctionTemplate(func(info v8.FunctionCallbackInfo) {
    data := new(Object)
    data._properties = make(map[string]interface{})
    // data._properties["style"] = "cacaface"
    // data._properties["test"] = 6
    info.This().SetInternalField(0, data)
  }, nil)
  ftConstructor.SetClassName("Object")
  obj_template := ftConstructor.InstanceTemplate()
  obj_template.SetNamedPropertyHandler(
    func(name string, info v8.PropertyCallbackInfo) {
      log.Printf("get %s\n", name)
      data := info.This().ToObject().GetInternalField(0).(*Object)
      cs := info.CurrentScope()

      if v, ok := data._properties[name]; ok {
        info.ReturnValue().Set(interfaceToValue(cs, v))
      } else {
        log.Printf("Object missing: %s\n", name)
        info.ReturnValue().Set(cs.GetEngine().Undefined())
      }
    },
    func(name string, value *v8.Value, info v8.PropertyCallbackInfo) {
      log.Printf("set %s\n", name)
      data := info.This().ToObject().GetInternalField(0).(*Object)
      data._properties[name] = valueToInterface(value)
      info.ReturnValue().Set(value)
    },
    func(name string, info v8.PropertyCallbackInfo) {
      log.Printf("query %s\n", name)
      info.ReturnValue().SetInt32(int32(v8.PA_None))
    },
    func(name string, info v8.PropertyCallbackInfo) {
      log.Printf("delete %s\n", name)
      data := info.This().ToObject().GetInternalField(0).(*Object)
      delete(data._properties, name)
    },
    func(info v8.PropertyCallbackInfo) {
      log.Printf("enumerate\n")
      cs := info.CurrentScope()
      data := info.This().ToObject().GetInternalField(0).(*Object)

      a := cs.NewArray(len(data._properties))
      x := 0
      for k, _ := range data._properties {
        a.SetElement(x, cs.NewString(k))
        x++
      }

      info.ReturnValue().Set(a.Value)
    },
    nil,
  )
  obj_template.SetInternalFieldCount(1)

  return obj_template, ftConstructor
}

func interfaceToValue(cs v8.ContextScope, v interface{}) *v8.Value {
  switch t := v.(type) {
  case string:
    return cs.NewString(t)
  case bool:
    return cs.NewBoolean(t)
  case int32:
    return cs.NewInteger(int64(t))
  case uint32:
    return cs.NewInteger(int64(t))
  case int64:
    return cs.NewInteger(t)
  case float64:
    return cs.NewNumber(t)
  case Undefined:
    return cs.GetEngine().Undefined()
  case map[string]interface{}:
    o_tpl, _ := NewObject(cs.GetEngine())
    o := o_tpl.NewObject()
    o2 := o.ToObject().GetInternalField(0).(*Object)
    o2._properties = t
    return o
  }

  rv := reflect.ValueOf(v)
  if rv.Type().Kind() == reflect.Array {
    a := cs.NewArray(rv.Len())
    for x := 0; x < rv.Len(); x++ {
      a.SetElement(x, interfaceToValue(cs, rv.Field(x).Interface()))
    }
  }

  return cs.GetEngine().Undefined()
}

func valueToInterface(value *v8.Value) interface{} {
  switch {
  case value.IsArray():
    a := value.ToArray()
    var ga []interface{}
    for x := 0; x < a.Length(); x++ {
      ga = append(ga, valueToInterface(a.GetElement(x)))
    }
    return ga
  case value.IsBoolean():
    return value.ToBoolean()
  case value.IsBooleanObject(): // fixme: handle differently?
    return value.ToBoolean()
  case value.IsDate():
  case value.IsExternal():
  case value.IsFalse():
    return value.IsFalse()
  case value.IsFunction():
    // value.ToFunction().
  case value.IsInt32():
    return value.ToInt32()
  case value.IsNativeError():
  case value.IsNull():
    return nil
  case value.IsNumber():
    return value.ToNumber()
  case value.IsNumberObject(): // fixme: handle differently?
    return value.ToNumber()
  case value.IsObject():
    return value.ToObject().GetInternalField(0)
  case value.IsRegExp():
  case value.IsString():
    return value.ToString()
  case value.IsStringObject(): // fixme: handle differently?
    return value.ToString()
  case value.IsTrue():
    return value.IsTrue()
  case value.IsUint32():
    return value.ToUint32()
  case value.IsUndefined():
    return Undefined{}
  }

  return nil
}

// obj_template, obj_constructor := NewObject(engine)
// document_createElementNS := engine.NewFunctionTemplate(func(info v8.FunctionCallbackInfo) {
//  log.Println("createElementNS")
//  o := obj_template.NewObject()
//  // o2 := o.ToObject().GetInternalField(0).(*Object)
//  // o2._properties["style"] = map[string]interface{}{
//  //  "webkitTapHighlightColor": "value",
//  // }
//  info.ReturnValue().Set(o)
// }, nil)

//cs.Global().SetProperty("Object", obj_constructor.NewFunction(), v8.PA_None)
//document := engine.NewObjectTemplate()
//document.SetProperty("createElementNS", document_createElementNS.NewFunction(), v8.PA_None)
//cs.Global().ForceSetProperty("document", document.NewObject(), v8.PA_ReadOnly|v8.PA_DontDelete)