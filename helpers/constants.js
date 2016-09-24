module.exports = {
	addressLength: 208,
	blockHeaderLength: 248,
	confirmationLength: 77,
	fees:{
		send: 10000000,
		vote: 100000000,
		secondsignature: 500000000,
		delegate: 2500000000,
		multisignature: 500000000,
		dapp: 2500000000
	},
	activeDelegates: 101,
	feeStart: 1,
	feeStartVolume: 10000 * 100000000,
	fixedPoint : Math.pow(10, 8),
	forgingTimeOut: 500, // 50 blocks
	maxAddressesLength: 208 * 128,
	maxAmount: 100000000,
	maxClientConnections: 100,
	maxConfirmations : 77 * 100,
	maxPayloadLength: 1024 * 1024,
	maxRequests: 10000 * 12,
	maxSignaturesLength: 196 * 256,
	maxTxsPerBlock: 25,
	numberLength: 100000000,
	requestLength: 104,
	rewards: {
		offset: 10,   // Start rewards at block (n)
		distance: 3000000, // Distance between each milestone
	},
	signatureExceptions: [
		"5676385569187187158", // 868797
		"5384302058030309746", // 869890
		"9352922026980330230", // 925165
	],
	signatureLength: 196,
	totalAmount: 1000000000000000,
	unconfirmedTransactionTimeOut: 10800, // 1080 blocks
	voteExceptions: [
		"5524930565698900323",  // 20407
		"11613486949732674475", // 123300
		"14164134775432642506"  // 123333
	]
}