// Teensy 4.1 motor/servo control
// Tilt: H-bridge  — PWM: pin 9, IN1: pin 6, IN2: pin 5
// Side: hobby servo — signal: pin 10
// Pon:  hobby servo — signal: pin 11
//
// Serial protocol: "pon,side,tilt\n"  (values -127..127)

#include <Servo.h>

#define PWM_TILT  9
#define IN1_TILT  6
#define IN2_TILT  5

#define PIN_SIDE  10
#define PIN_PON   24

Servo servoSide;
Servo servoPon;


// H-bridge: vel -127..127
void applyTilt(int vel) {
  vel = constrain(vel, -127, 127);
  if (vel > 0) {
    digitalWrite(IN1_TILT, HIGH);
    digitalWrite(IN2_TILT, LOW);
    analogWrite(PWM_TILT, map(vel, 0, 127, 0, 255));
  } else if (vel < 0) {
    digitalWrite(IN1_TILT, LOW);
    digitalWrite(IN2_TILT, HIGH);
    analogWrite(PWM_TILT, map(-vel, 0, 127, 0, 255));
  } else {
    digitalWrite(IN1_TILT, LOW);
    digitalWrite(IN2_TILT, LOW);
    analogWrite(PWM_TILT, 0);
  }
}

// Hobby servo: vel -127..127 → 1000..2000 µs
void applyServo(Servo &s, int vel) {
  vel = constrain(vel, -127, 127);
  vel = map(vel, -127, 127, 1000, 2000);
  servoSide.writeMicroseconds(vel);

  s.write(vel);
}

void setup() {
  Serial.begin(115200);
  while (!Serial && millis() < 1000);
  Serial.println("Ready. Expecting: pon,side,tilt\\n  (-127..127)");

  pinMode(PWM_TILT, OUTPUT);
  pinMode(IN1_TILT, OUTPUT);
  pinMode(IN2_TILT, OUTPUT);
  analogWriteResolution(8);

  servoSide.attach(PIN_SIDE);
  servoPon.attach(PIN_PON);

  servoSide.writeMicroseconds(1500);


  applyTilt(0);
  applyServo(servoSide, 0);
  applyServo(servoPon, 0);
}

char incomingByte = 0;

int side_current = 0;
int pon_current = 0;
int tilt_current = 0;


int side_target = 0;
int tilt_target = 0;

void loop() {

  if (Serial.available()) {

    pon_current = Serial.parseInt();
    side_target  = Serial.parseInt();
    tilt_target = Serial.parseInt();
  
    
    while (Serial.available()){
      Serial.read();
    }

    Serial.print("pon: ");
    Serial.print(pon_current);
    Serial.print("side: ");
    Serial.print(side_target);
    Serial.print("tilt: ");
    Serial.print(tilt_current);
    Serial.println();

  }
  
  side_current = 0.9*side_current + 0.1*side_target;
  tilt_current = 0.8*tilt_current + 0.2*tilt_target;

  applyServo(servoPon,  pon_current);
  applyServo(servoSide, side_current);
  applyTilt(tilt_current);
  delay(10);
}
