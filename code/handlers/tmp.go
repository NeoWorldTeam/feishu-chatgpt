// handlerType := judgeChatType(event)
// 	if handlerType == "otherChat" {
// 		fmt.Println("unknown chat type")
// 		return nil
// 	}
// 	msgType := judgeMsgType(event)
// 	if msgType != "text" {
// 		fmt.Println("msg type:", msgType, *event.Event.Message.Content)
// 		// return nil
// 	}
// 	if msgType == "image" {
// 		image := &larkim.CreateImageRespData{}
// 		json.Unmarshal([]byte(*event.Event.Message.Content), image)
// 		req := larkim.NewGetImageReqBuilder().
// 			ImageKey(*image.ImageKey).
// 			Build()
// 		resp, err := initialization.GetLarkClient().Im.Image.Get(ctx, req)
// 		// resp.WriteFile(*image.ImageKey)
// 		// fmt.Println("base64 image str:", base64.RawStdEncoding.EncodeToString(resp.RawBody))
// 		fmt.Println("get image by key", *image.ImageKey, larkcore.Prettify(resp), err)
// 		return nil
// 	}
// 	return handlers[handlerType].handle(ctx, event)