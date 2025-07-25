import random 

byte_data = bytearray()

for i in range(160):
    for j in range(144):
        pixel_value = random.randint(0, 3)
        
        # Convert to 2-bit binary representation
        if pixel_value == 0:    # 00
            byte_data.append(0b00)
        elif pixel_value == 1:  # 01
            byte_data.append(0b01)
        elif pixel_value == 2:  # 10
            byte_data.append(0b10)
        else:                   # 11
            byte_data.append(0b11)

with open("sample.bin","wb") as f:
    f.write(byte_data)