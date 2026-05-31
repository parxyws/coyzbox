package helper

import (
	"fmt"
	"strings"

	"github.com/parxyws/cozybox/internal/models"
)

const (
	ID_PREFIX_QUOTATION      = "QUO"
	ID_PREFIX_INVOICE        = "INV"
	ID_PREFIX_RECEIPT        = "REC"
	ID_PREFIX_PURCHASE_ORDER = "PO"
	ID_PREFIX_SALES_ORDER    = "SO"
	ID_PREFIX_DEBIT_NOTE     = "DN"
)

func getLastId() string {

	return ""
}

func IdentificationGenerator(uniqueId string, documentType models.DocumentType) string {
	var documentTypeId string
	lastId := strings.Split(getLastId(), "/")

	switch documentType {
	case models.DocumentTypeQuotation:
		documentTypeId = ID_PREFIX_QUOTATION
	case models.DocumentTypeInvoice:
		documentTypeId = ID_PREFIX_INVOICE
	case models.DocumentTypeReceipt:
		documentTypeId = ID_PREFIX_RECEIPT
	case models.DocumentTypePurchaseOrder:
		documentTypeId = ID_PREFIX_PURCHASE_ORDER
	case models.DocumentTypeSalesOrder:
		documentTypeId = ID_PREFIX_SALES_ORDER
	case models.DocumentTypeDebitNote:
		documentTypeId = ID_PREFIX_DEBIT_NOTE
	}

	generatedId := fmt.Sprintf("%s-%s-%s", documentTypeId, uniqueId, lastId)

	return generatedId
}
