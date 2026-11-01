import "dotenv/config";
import { PrismaPg } from "@prisma/adapter-pg";
import bcrypt from "bcryptjs";
import { PrismaClient } from "../generated/prisma/client";

const adapter = new PrismaPg({ connectionString: process.env.DATABASE_URL! });
const prisma = new PrismaClient({ adapter });

const ROLES = ["administrator", "manajer", "partner", "customer"];

async function main() {
  const email = process.env.ADMIN_EMAIL;
  const name = process.env.ADMIN_NAME ?? "Administrator";
  const password = process.env.ADMIN_PASSWORD;

  if (!email || !password) {
    throw new Error("ADMIN_EMAIL dan ADMIN_PASSWORD wajib diisi di .env");
  }
  if (
    password.length < 8 ||
    !/[A-Za-z]/.test(password) ||
    !/\d/.test(password) ||
    password.toLowerCase() === email.toLowerCase()
  ) {
    throw new Error("Kata sandi admin harus minimal 8 karakter, berisi huruf dan angka, dan tidak sama dengan email");
  }

  for (const roleName of ROLES) {
    await prisma.role.upsert({
      where: { name: roleName },
      update: {},
      create: { name: roleName },
    });
  }

  const adminRole = await prisma.role.findUniqueOrThrow({
    where: { name: "administrator" },
  });

  // cost 10 = bcrypt.DefaultCost di backend Go
  const passwordHash = await bcrypt.hash(password, 10);

  await prisma.user.upsert({
    where: { email },
    update: {}, // akun yang sudah ada tidak ditimpa
    create: {
      email,
      passwordHash,
      fullName: name,
      roleId: adminRole.id,
      isActive: true,
    },
  });

  console.log("Seeder selesai: role dan akun admin awal siap.");
}

main()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });