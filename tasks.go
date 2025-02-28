package bilicoin

import (
	"errors"
	"time"
)

var TaskMap = map[string]Task{
	"drop-coin":     TaskDropCoin,
	"silver-2-coin": TaskSilver2Coin,
}

func TaskSilver2Coin(user *BiliUser) error {
	// 获取用户信息失败
	if err := user.GetBiliWallet(); err != nil {
		return err
	}
	// 使用银瓜子兑换硬币一枚
	if user.Bi.Silver >= 700 {
		return user.Silver2Coin()
	}
	return errors.New("not enough silver")
}

const maxLoopCount = 5 + 3 // only three failures are allowed...

func TaskDropCoin(user *BiliUser) error {
	// 获取日志失败
	// 过期、未知错误、服务不可达、解析错误
	if err := user.GetBiliCoinLog(); err != nil {
		return err
	}
	for loopCount := 0; true; loopCount++ {
		//if user.DropCoinCount > 4 {
		//	user.InfoUpdate()
		//	return nil
		//}
		user.RandDrop()
		ra := Random(10)
		time.Sleep(ra)
	}
	user.InfoUpdate()
	return errors.New("max retries reached")
}
