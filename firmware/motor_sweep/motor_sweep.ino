// H-Bridge serial control for Teensy 4.1
// Enable: pin 9 (PWM), IN1: pin 6, IN2: pin 5
// Receives signed integer speed (-255..255) over serial, newline terminated

#define EN  9
#define IN1 6
#define IN2 5

void applySpeed(int spd) {
  spd = constrain(spd, -255, 255);
  if (spd > 0) {
    digitalWrite(IN1, HIGH);
    digitalWrite(IN2, LOW);
    analogWrite(EN, spd);
  } else if (spd < 0) {
    digitalWrite(IN1, LOW);
    digitalWrite(IN2, HIGH);
    analogWrite(EN, -spd);
  } else {
    analogWrite(EN, 0);
    digitalWrite(IN1, LOW);
    digitalWrite(IN2, LOW);
  }
}

void setup() {
  Serial.begin(115200);
  while (!Serial && millis() < 2000);
  Serial.println("Ready. Waiting for speed values (-255..255).");

  pinMode(EN,  OUTPUT);
  pinMode(IN1, OUTPUT);
  pinMode(IN2, OUTPUT);
  analogWriteResolution(8);
  applySpeed(0);
}

void loop() {
  if (Serial.available()) {
    int spd = Serial.parseInt();
    applySpeed(spd);
  }
}
