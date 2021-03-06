package main

type Order struct {
	ObjectType 			string 	`json:"docType"`
	OrderId      		string 	`json:"orderId"`
	FromAddress     	string 	`json:"fromAddress"`
	ToAddress      		string 	`json:"toAddress"`
	Content      		string 	`json:"content"`
	WeightTon      		float64 `json:"weightTon"`
	TransFee      		float64 `json:"transFee"`
	OrderState      	string 	`json:"orderState"` //in [WAIT_DRIVER_ACCEPT, DRIVER_ACCEPT_WAIT_ROAD, DRIVER_ON_ROAD, ARRIVED_WAIT_SIGN, SIGNED]
	GoodsOwnerId    	string 	`json:"goodsOwnerId"`
	BrokerId      		string 	`json:"brokerId"`
	DriverId      		string 	`json:"driverId"`
	CreateDate      	string	`json:"createDate"`
	Open				bool	`json:"open"`
  
	ChangeStateHistory map[string]string
  }

type UpdatePositionHistory struct {
	ObjectType 			string  `json:"docType"`
	PositionId			string  `json:"positionId"`
	OrderId      		string  `json:"orderId"`
	Sequence			string  `json:"sequence"`
	TimePosition		string  `json:"timePosition"`
	PositionString		string  `json:"positionString"`
} 

type StringHash struct {
	ObjectType 			string  `json:"docType"`
	DataId				string  `json:"dataId"`
	OrderId      		string  `json:"orderId"`
	DataUrl				string  `json:"dataUrl"`
	ShaResult			string  `json:"shaResult"`
	Comment				string  `json:"comment"`
} 

type FileHash struct {
	ObjectType 			string  `json:"docType"`
	FileId      		string  `json:"fileId"`
	OrderId				string  `json:"orderId"`
	DataUrl				string  `json:"dataUrl"`
	ShaResult			string  `json:"shaResult"`
	Comment				string  `json:"comment"`
} 

type StringWithKey struct {
	Key 				string  				`json:"Key"`
	Record				StringHash   			`json:"Record"`
}

type FileWithKey struct {
	Key 				string  				`json:"Key"`
	Record				FileHash   			`json:"Record"`
}

type User struct {
	ObjectType 			string  				`json:"docType"`
	UserId				string					`json:"userId"`
	UserName			string					`json:"userName"`
	Role				string					`json:"role"`
	Telephone			string					`json:"telephone"`
	Valid				bool					`json:"valid"`
}

type UserGenerated struct {
	StatusMessage 		string       			`json:"statusMessage"`
	User				User					`json:"user"`
	File       			[]FileHash 				`json:"file"`
}

type AutoGenerated struct {
	StatusMessage 		string       			`json:"statusMessage"`
	Order				Order					`json:"order"`
	String       		[]StringHash 			`json:"string"`
	File       			[]FileHash 				`json:"file"`
	Position       		[]UpdatePositionHistory	`json:"position"`
}
