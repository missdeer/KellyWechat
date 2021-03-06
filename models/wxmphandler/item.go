package models

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/missdeer/KellyWechat/models/wd"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	maxArrayLength int = 10
)

func ItemId(req *Request, resp *Response) error {
	userInputText := strings.Trim(strings.ToLower(req.Content), " ")
	uuid, err := strconv.ParseUint(userInputText, 10, 64)
	if err != nil {
		beego.Error("incorrect input ", userInputText)
		resp.Content = "衣丽已经很努力地在学习了，但仍然不能理解您的需求，请您输入help查看衣丽能懂的一些命令吧:("
		return nil
	}
	item := &models.WDItem{}
	item.Uuid = uuid
	if item.Get("uuid") != nil {
		beego.Error("not found ", userInputText)

		qs := models.Items()
		var items []models.WDItem
		n, err := qs.Limit(500).All(&items)
		if err != nil || n == 0 {
			resp.Content = fmt.Sprintf("好像找不到编号为%s的宝贝哦:(", userInputText)
			return nil
		}
		wdItem := items[rand.Intn(int(n))]

		var a WXMPItem
		resp.MsgType = News
		resp.ArticleCount = 1
		a.Description = wdItem.Name + `，点击查看详细信息哦:)`
		a.Title = fmt.Sprintf("好像找不到编号为%s的宝贝哦，随便看点东西吧:)", userInputText)
		a.PicUrl = wdItem.Logo + "?w=360&h=0"
		a.Url = fmt.Sprintf(`http://wd.koudai.com/i/%d`, wdItem.Uuid)
		resp.Articles = append(resp.Articles, &a)
		resp.FuncFlag = 1
		return nil
	}

	var a WXMPItem
	resp.MsgType = News
	resp.ArticleCount = 1
	a.Description = `点击查看详细信息哦:)`
	a.Title = item.Name
	a.PicUrl = item.Logo + "?w=360&h=0"
	a.Url = fmt.Sprintf(`http://wd.koudai.com/i/%d`, item.Uuid)
	resp.Articles = append(resp.Articles, &a)
	resp.FuncFlag = 1
	return nil
}

func composeItemListReponse(items []models.WDItem, shopId int, req *Request, resp *Response) error {
	n := len(items)
	resp.MsgType = News
	resp.ArticleCount = n
	arrayLength := n
	if arrayLength < maxArrayLength-1 {
		arrayLength++
	} else {
		arrayLength = maxArrayLength
	}
	a := make([]WXMPItem, arrayLength)
	rand.Seed(time.Now().UnixNano())
	for i, index := range rand.Perm(n) {
		if n > maxArrayLength {
			if shopId >= 0 && i >= maxArrayLength-1 {
				resp.ArticleCount = maxArrayLength - 1
				break
			}
			if shopId < 0 && i >= maxArrayLength {
				resp.ArticleCount = maxArrayLength
				break
			}
		}
		item := items[index]
		a[i].Description = `点击查看详细信息哦:)`
		a[i].Title = item.Name
		a[i].PicUrl = item.Logo + "?w=360&h=0"
		a[i].Url = fmt.Sprintf(`http://wd.koudai.com/i/%d`, item.Uuid)
		resp.Articles = append(resp.Articles, &a[i])
	}

	if n > maxArrayLength && shopId >= 0 {
		wdShop := &models.WDShop{}
		wdShop.Id = shopId
		shopItem := &WXMPItem{}
		if wdShop.Get("id") == nil {
			resp.ArticleCount++
			shopItem.Description = wdShop.Note
			shopItem.Title = `宝贝数量较多，请进入微店查看更多 - ` + wdShop.Name
			shopItem.PicUrl = wdShop.Logo + "?w=360&h=0"
			shopItem.Url = fmt.Sprintf(`http://wd.koudai.com/s/%d`, wdShop.Uuid)
			resp.Articles = append(resp.Articles, shopItem)
		}
	}
	resp.FuncFlag = 1
	return nil
}

func ItemListByShopUuid(uuid string, req *Request, resp *Response) error {
	wdShop := &models.WDShop{}
	var err error
	wdShop.Uuid, err = strconv.ParseUint(uuid, 10, 64)
	if err != nil {
		beego.Error("can't convert uuid: ", uuid, err)
		return err
	}
	if err = wdShop.Get("uuid"); err != nil {
		beego.Error("can't find this shop:", uuid, wdShop)
		return err
	}
	return itemList(wdShop.Id, req, resp)
}

func itemList(shopId int, req *Request, resp *Response) error {
	qs := models.Items()
	var items []models.WDItem
	n, err := qs.Limit(500).Filter("shop_id", shopId).All(&items)
	if err != nil || n == 0 {
		resp.Content = `没开这样的店哦:(`
		return nil
	}

	return composeItemListReponse(items, shopId, req, resp)
}

func SearchItems(req *Request, resp *Response) error {
	userInputText := strings.Trim(strings.ToLower(req.Content), " ")
	qs := models.Items()
	var items []models.WDItem
	n, err := qs.Limit(500).Filter("name__icontains", userInputText).All(&items)
	if err != nil || n == 0 {
		qs := models.Items()
		var items []models.WDItem
		n, err := qs.Limit(500).All(&items)
		if err != nil || n == 0 {
			resp.Content = fmt.Sprintf("好像找不到包含关键字“%s”的宝贝哦:(", userInputText)
			return nil
		}
		wdItem := items[rand.Intn(int(n))]

		var a WXMPItem
		resp.MsgType = News
		resp.ArticleCount = 1
		a.Description = wdItem.Name + `，点击查看详细信息哦:)`
		a.Title = fmt.Sprintf("好像找不到包含关键字“%s”的宝贝哦，随便看点东西吧:)", userInputText)
		a.PicUrl = wdItem.Logo + "?w=360&h=0"
		a.Url = fmt.Sprintf(`http://wd.koudai.com/i/%d`, wdItem.Uuid)
		resp.Articles = append(resp.Articles, &a)
		resp.FuncFlag = 1

		return nil
	}

	return composeItemListReponse(items, -1, req, resp)
}
