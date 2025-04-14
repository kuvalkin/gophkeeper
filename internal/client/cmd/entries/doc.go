/*
Package entries provides structures and methods for handling various types of
entries, such as login credentials and bank card details. It includes
serialization and deserialization logic for these entries using Protocol Buffers
and ensures data validation.

The package currently supports the following entry types:
  - LoginPasswordPair: Represents a pair of login credentials with a login and password.
  - BankCard: Represents a bank card with details such as number, holder name, CVV, and expiration date.

Each entry type provides methods for:
  - Marshal: Serializing the entry into a Protocol Buffers format.
  - Unmarshal: Deserializing the entry from a Protocol Buffers format.

Validation is performed during the marshaling process to ensure the integrity
and correctness of the data. For example:
  - LoginPasswordPair validates the presence of login and password fields.
  - BankCard validates the card number using the Luhn algorithm and checks
    the expiration date format.
*/
package entries
