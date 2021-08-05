package important

func Mobile(str string) string {
	if len(str) == 11 {
		return str[0:3] + "****" + str[7:]
	} else {
		return str
	}
}

func BankCard(str string) string {
	if len(str) == 16 || len(str) == 19 {
		return str[0:8] + "*******" + str[15:]
	} else {
		return str
	}
}

func IdCard(str string) string {
	if len(str) == 15 || len(str) == 18 {
		return str[0:4] + "**********" + str[14:]
	} else {
		return str
	}
}

func Token(str string) string {
	if len(str) > 5 {
		return str[0:4] + "****" + str[len(str)-5:]
	}else{
		return str
	}
}
